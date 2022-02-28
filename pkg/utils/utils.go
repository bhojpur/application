package utils

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"bufio"
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	appsvr "github.com/bhojpur/application/pkg/engine"
	orm "github.com/bhojpur/orm/pkg/engine"
	now "github.com/bhojpur/orm/pkg/now"
	"github.com/gosimple/slug"
	"github.com/microcosm-cc/bluemonday"

	"github.com/docker/docker/client"
	"github.com/gocarina/gocsv"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v2"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	socketFormat = "%s/app-%s-%s.socket"
)

var (
	clientSet  *kubernetes.Clientset
	kubeConfig *rest.Config

	envRegexp = regexp.MustCompile(`(?m)(,)\s*[a-zA-Z\_][a-zA-Z0-9\_]*=`)
)

// AppRoot app root path
var AppRoot, _ = os.Getwd()

// ContextKey defined type used for context's key
type ContextKey string

// ContextDBName db name used for context
var ContextDBName ContextKey = "ContextDB"

// HTMLSanitizer html sanitizer to avoid XSS
var HTMLSanitizer = bluemonday.UGCPolicy()

func init() {
	HTMLSanitizer.AllowStandardAttributes()
	if path := os.Getenv("WEB_ROOT"); path != "" {
		AppRoot = path
	}
}

// GOPATH return GOPATH from env
func GOPATH() []string {
	paths := strings.Split(os.Getenv("GOPATH"), string(os.PathListSeparator))
	if len(paths) == 0 {
		fmt.Println("GOPATH doesn't exist")
	}
	return paths
}

// GetDBFromRequest get database from request
var GetDBFromRequest = func(req *http.Request) *orm.DB {
	db := req.Context().Value(ContextDBName)
	if tx, ok := db.(*orm.DB); ok {
		return tx
	}

	return nil
}

// HumanizeString Humanize separates string based on capitalizd letters
// e.g. "OrderItem" -> "Order Item"
func HumanizeString(str string) string {
	var human []rune
	for i, l := range str {
		if i > 0 && isUppercase(byte(l)) {
			if (!isUppercase(str[i-1]) && str[i-1] != ' ') || (i+1 < len(str) && !isUppercase(str[i+1]) && str[i+1] != ' ' && str[i-1] != ' ') {
				human = append(human, rune(' '))
			}
		}
		human = append(human, l)
	}
	return strings.Title(string(human))
}

func isUppercase(char byte) bool {
	return 'A' <= char && char <= 'Z'
}

var asicsiiRegexp = regexp.MustCompile("^(\\w|\\s|-|!)*$")

// ToParamString replaces spaces and separates words (by uppercase letters) with
// underscores in a string, also downcase it
// e.g. ToParamString -> to_param_string, To ParamString -> to_param_string
func ToParamString(str string) string {
	if asicsiiRegexp.MatchString(str) {
		return orm.ToDBName(strings.Replace(str, " ", "_", -1))
	}
	return slug.Make(str)
}

// PatchURL updates the query part of the request url.
//     PatchURL("google.com","key","value") => "google.com?key=value"
func PatchURL(originalURL string, params ...interface{}) (patchedURL string, err error) {
	url, err := url.Parse(originalURL)
	if err != nil {
		return
	}

	query := url.Query()
	for i := 0; i < len(params)/2; i++ {
		// Check if params is key&value pair
		key := fmt.Sprintf("%v", params[i*2])
		value := fmt.Sprintf("%v", params[i*2+1])

		if value == "" {
			query.Del(key)
		} else {
			query.Set(key, value)
		}
	}

	url.RawQuery = query.Encode()
	patchedURL = url.String()
	return
}

// JoinURL updates the path part of the request url.
//     JoinURL("google.com", "admin") => "google.com/admin"
//     JoinURL("google.com?q=keyword", "admin") => "google.com/admin?q=keyword"
func JoinURL(originalURL string, paths ...interface{}) (joinedURL string, err error) {
	u, err := url.Parse(originalURL)
	if err != nil {
		return
	}

	var urlPaths = []string{u.Path}
	for _, p := range paths {
		urlPaths = append(urlPaths, fmt.Sprint(p))
	}

	if strings.HasSuffix(strings.Join(urlPaths, ""), "/") {
		u.Path = path.Join(urlPaths...) + "/"
	} else {
		u.Path = path.Join(urlPaths...)
	}

	joinedURL = u.String()
	return
}

// SetCookie set cookie for context
func SetCookie(cookie http.Cookie, context *appsvr.Context) {
	cookie.HttpOnly = true

	// set https cookie
	if context.Request != nil && context.Request.URL.Scheme == "https" {
		cookie.Secure = true
	}

	// set default path
	if cookie.Path == "" {
		cookie.Path = "/"
	}

	http.SetCookie(context.Writer, &cookie)
}

// Stringify stringify any data, if it is a struct, will try to use its Name, Title, Code field, else will use its primary key
func Stringify(object interface{}) string {
	if obj, ok := object.(interface {
		Stringify() string
	}); ok {
		return obj.Stringify()
	}

	scope := orm.Scope{Value: object}
	for _, column := range []string{"Name", "Title", "Code"} {
		if field, ok := scope.FieldByName(column); ok {
			if field.Field.IsValid() {
				result := field.Field.Interface()
				if valuer, ok := result.(driver.Valuer); ok {
					if result, err := valuer.Value(); err == nil {
						return fmt.Sprint(result)
					}
				}
				return fmt.Sprint(result)
			}
		}
	}

	if scope.PrimaryField() != nil {
		if scope.PrimaryKeyZero() {
			return ""
		}
		return fmt.Sprintf("%v#%v", scope.GetModelStruct().ModelType.Name(), scope.PrimaryKeyValue())
	}

	return fmt.Sprint(reflect.Indirect(reflect.ValueOf(object)).Interface())
}

// ModelType get value's model type
func ModelType(value interface{}) reflect.Type {
	reflectType := reflect.Indirect(reflect.ValueOf(value)).Type()

	for reflectType.Kind() == reflect.Ptr || reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
	}

	return reflectType
}

// ParseTagOption parse tag options to hash
func ParseTagOption(str string) map[string]string {
	tags := strings.Split(str, ";")
	setting := map[string]string{}
	for _, value := range tags {
		v := strings.Split(value, ":")
		k := strings.TrimSpace(strings.ToUpper(v[0]))
		if len(v) == 2 {
			setting[k] = v[1]
		} else {
			setting[k] = k
		}
	}
	return setting
}

// ExitWithMsg debug error messages and print stack
func ExitWithMsg(msg interface{}, value ...interface{}) {
	fmt.Printf("\n"+filenameWithLineNum()+"\n"+fmt.Sprint(msg)+"\n", value...)
	debug.PrintStack()
}

// FileServer file server that disabled file listing
func FileServer(dir http.Dir) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := path.Join(string(dir), r.URL.Path)
		if f, err := os.Stat(p); err == nil && !f.IsDir() {
			http.ServeFile(w, r, p)
			return
		}

		http.NotFound(w, r)
	})
}

func filenameWithLineNum() string {
	var total = 10
	var results []string
	for i := 2; i < 15; i++ {
		if _, file, line, ok := runtime.Caller(i); ok {
			total--
			results = append(results[:0],
				append(
					[]string{fmt.Sprintf("%v:%v", strings.TrimPrefix(file, os.Getenv("GOPATH")+"src/"), line)},
					results[0:]...)...)

			if total == 0 {
				return strings.Join(results, "\n")
			}
		}
	}
	return ""
}

// GetLocale get locale from request, cookie, after get the locale, will write the locale to the cookie if possible
// Overwrite the default logic with
//     utils.GetLocale = func(context *appsvr.Context) string {
//         // ....
//     }
var GetLocale = func(context *appsvr.Context) string {
	if locale := context.Request.Header.Get("Locale"); locale != "" {
		return locale
	}

	if locale := context.Request.URL.Query().Get("locale"); locale != "" {
		if context.Writer != nil {
			context.Request.Header.Set("Locale", locale)
			SetCookie(http.Cookie{Name: "locale", Value: locale, Expires: time.Now().AddDate(1, 0, 0)}, context)
		}
		return locale
	}

	if locale, err := context.Request.Cookie("locale"); err == nil {
		return locale.Value
	}

	return ""
}

// ParseTime parse time from string
// Overwrite the default logic with
//     utils.ParseTime = func(timeStr string, context *appsvr.Context) (time.Time, error) {
//         // ....
//     }
var ParseTime = func(timeStr string, context *appsvr.Context) (time.Time, error) {
	return now.Parse(timeStr)
}

// FormatTime format time to string
// Overwrite the default logic with
//     utils.FormatTime = func(time time.Time, format string, context *appsvr.Context) string {
//         // ....
//     }
var FormatTime = func(date time.Time, format string, context *appsvr.Context) string {
	return date.Format(format)
}

var replaceIdxRegexp = regexp.MustCompile(`\[\d+\]`)

// SortFormKeys sort form keys
func SortFormKeys(strs []string) {
	sort.Slice(strs, func(i, j int) bool { // true for first
		str1 := strs[i]
		str2 := strs[j]
		matched1 := replaceIdxRegexp.FindAllStringIndex(str1, -1)
		matched2 := replaceIdxRegexp.FindAllStringIndex(str2, -1)

		for x := 0; x < len(matched1); x++ {
			prefix1 := str1[:matched1[x][0]]
			prefix2 := str2

			if len(matched2) >= x+1 {
				prefix2 = str2[:matched2[x][0]]
			}

			if prefix1 != prefix2 {
				return strings.Compare(prefix1, prefix2) < 0
			}

			if len(matched2) < x+1 {
				return false
			}

			number1 := str1[matched1[x][0]:matched1[x][1]]
			number2 := str2[matched2[x][0]:matched2[x][1]]

			if number1 != number2 {
				if len(number1) != len(number2) {
					return len(number1) < len(number2)
				}
				return strings.Compare(number1, number2) < 0
			}
		}

		return strings.Compare(str1, str2) < 0
	})
}

// GetAbsURL get absolute URL from request
func GetAbsURL(req *http.Request) url.URL {
	if req.URL.IsAbs() {
		return *req.URL
	}

	var result *url.URL
	if domain := req.Header.Get("Origin"); domain != "" {
		result, _ = url.Parse(domain)
	} else {
		if req.TLS == nil {
			result, _ = url.Parse("http://" + req.Host)
		} else {
			result, _ = url.Parse("https://" + req.Host)
		}
	}

	result.Parse(req.RequestURI)
	return *result
}

// Indirect returns last value that v points to
func Indirect(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}
	return v
}

// SliceUniq removes duplicate values in given slice
func SliceUniq(s []string) []string {
	for i := 0; i < len(s); i++ {
		for i2 := i + 1; i2 < len(s); i2++ {
			if s[i] == s[i2] {
				// delete
				s = append(s[:i2], s[i2+1:]...)
				i2--
			}
		}
	}
	return s
}

// SafeJoin safe join https://snyk.io/research/zip-slip-vulnerability#go
func SafeJoin(paths ...string) (string, error) {
	result := path.Join(paths...)
	// check filepath
	if !strings.HasPrefix(result, filepath.Clean(paths[0])+string(os.PathSeparator)) {
		return "", errors.New("invalid filepath")
	}

	return result, nil
}

func initKubeConfig() {
	kubeConfig = GetConfig()
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		panic(err)
	}

	clientSet = clientset
}

// GetConfig gets a kubernetes rest config.
func GetConfig() *rest.Config {
	if kubeConfig != nil {
		return kubeConfig
	}

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	conf, err := rest.InClusterConfig()
	if err != nil {
		conf, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err)
		}
	}

	return conf
}

// GetKubeClient gets a kubernetes client.
func GetKubeClient() *kubernetes.Clientset {
	if clientSet == nil {
		initKubeConfig()
	}

	return clientSet
}

// ToISO8601DateTimeString converts dateTime to ISO8601 Format
// ISO8601 Format: 2020-01-01T01:01:01.10101Z.
func ToISO8601DateTimeString(dateTime time.Time) string {
	return dateTime.UTC().Format("2006-01-02T15:04:05.999999Z")
}

// add env-vars from annotations.
func ParseEnvString(envStr string) []corev1.EnvVar {
	indexes := envRegexp.FindAllStringIndex(envStr, -1)
	lastEnd := len(envStr)
	parts := make([]string, len(indexes)+1)
	for i := len(indexes) - 1; i >= 0; i-- {
		parts[i+1] = strings.TrimSpace(envStr[indexes[i][0]+1 : lastEnd])
		lastEnd = indexes[i][0]
	}
	parts[0] = envStr[0:lastEnd]

	envVars := make([]corev1.EnvVar, 0)
	for _, s := range parts {
		pairs := strings.Split(strings.TrimSpace(s), "=")
		if len(pairs) != 2 {
			continue
		}
		envVars = append(envVars, corev1.EnvVar{
			Name:  pairs[0],
			Value: pairs[1],
		})
	}

	return envVars
}

// StringSliceContains return true if an array containe the "str" string.
func StringSliceContains(needle string, haystack []string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

// PrintTable to print in the table format.
func PrintTable(csvContent string) {
	WriteTable(os.Stdout, csvContent)
}

// WriteTable writes the csv table to writer.
func WriteTable(writer io.Writer, csvContent string) {
	table := tablewriter.NewWriter(writer)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetRowLine(false)
	table.SetCenterSeparator("")
	table.SetRowSeparator("")
	table.SetColumnSeparator("")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	scanner := bufio.NewScanner(strings.NewReader(csvContent))
	header := true

	for scanner.Scan() {
		text := strings.Split(scanner.Text(), ",")

		if header {
			table.SetHeader(text)
			header = false
		} else {
			table.Append(text)
		}
	}

	table.Render()
}

func TruncateString(str string, maxLength int) string {
	strLength := len(str)
	if strLength <= maxLength {
		return str
	}

	return str[0:maxLength-3] + "..."
}

func RunCmdAndWait(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	err = cmd.Start()
	if err != nil {
		return "", err
	}

	resp, err := ioutil.ReadAll(stdout)
	if err != nil {
		return "", err
	}
	errB, err := ioutil.ReadAll(stderr)
	if err != nil {
		return "", nil
	}

	err = cmd.Wait()
	if err != nil {
		// in case of error, capture the exact message
		if len(errB) > 0 {
			return "", errors.New(string(errB))
		}
		return "", err
	}

	return string(resp), nil
}

func CreateContainerName(serviceContainerName string, dockerNetwork string) string {
	if dockerNetwork != "" {
		return fmt.Sprintf("%s_%s", serviceContainerName, dockerNetwork)
	}

	return serviceContainerName
}

func CreateDirectory(dir string) error {
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		return nil
	}
	return os.Mkdir(dir, 0777)
}

// IsDockerInstalled checks whether docker is installed/running.
func IsDockerInstalled() bool {
	// nolint:staticcheck
	cli, err := client.NewEnvClient()
	if err != nil {
		return false
	}
	_, err = cli.Ping(context.Background())
	return err == nil
}

// IsAppListeningOnPort checks if Bhojpur Application is litening to a given port.
func IsAppListeningOnPort(port int, timeout time.Duration) error {
	start := time.Now()
	for {
		host := fmt.Sprintf("127.0.0.1:%v", port)
		conn, err := net.DialTimeout("tcp", host, timeout)
		if err == nil {
			conn.Close()
			return nil
		}

		if time.Since(start).Seconds() >= timeout.Seconds() {
			// Give up.
			return err
		}

		time.Sleep(time.Second)
	}
}

func IsAppListeningOnSocket(socket string, timeout time.Duration) error {
	start := time.Now()
	for {
		conn, err := net.DialTimeout("unix", socket, timeout)
		if err == nil {
			conn.Close()
			return nil
		}

		if time.Since(start).Seconds() >= timeout.Seconds() {
			// Give up.
			return err
		}

		time.Sleep(time.Second)
	}
}

func MarshalAndWriteTable(writer io.Writer, in interface{}) error {
	table, err := gocsv.MarshalString(in)
	if err != nil {
		return err
	}

	WriteTable(writer, table)
	return nil
}

func PrintDetail(writer io.Writer, outputFormat string, list interface{}) error {
	obj := list
	s := reflect.ValueOf(list)
	if s.Kind() == reflect.Slice && s.Len() == 1 {
		obj = s.Index(0).Interface()
	}

	var err error
	output := []byte{}

	switch outputFormat {
	case "yaml":
		output, err = yaml.Marshal(obj)
	case "json":
		output, err = json.MarshalIndent(obj, "", "  ")
	}
	if err != nil {
		return err
	}

	_, err = writer.Write(output)
	return err
}

func IsAddressLegal(address string) bool {
	var isLegal bool
	if address == "localhost" {
		isLegal = true
	} else if net.ParseIP(address) != nil {
		isLegal = true
	}
	return isLegal
}

func GetSocket(path, appID, protocol string) string {
	return fmt.Sprintf(socketFormat, path, appID, protocol)
}
