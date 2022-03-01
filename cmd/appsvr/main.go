package main

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
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/valyala/fasthttp"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/bhojpur/application/pkg/runtime"
	"github.com/bhojpur/service/pkg/utils/logger"

	// Included components in compiled Bhojpur Application runtime engine.

	// Secret stores.
	"github.com/bhojpur/service/pkg/secretstores"
	alicloud_paramstore "github.com/bhojpur/service/pkg/secretstores/alicloud/parameterstore"
	"github.com/bhojpur/service/pkg/secretstores/aws/parameterstore"
	"github.com/bhojpur/service/pkg/secretstores/aws/secretmanager"
	"github.com/bhojpur/service/pkg/secretstores/azure/keyvault"
	gcp_secretmanager "github.com/bhojpur/service/pkg/secretstores/gcp/secretmanager"
	"github.com/bhojpur/service/pkg/secretstores/hashicorp/vault"
	secretstore_kubernetes "github.com/bhojpur/service/pkg/secretstores/kubernetes"
	secretstore_env "github.com/bhojpur/service/pkg/secretstores/local/env"
	secretstore_file "github.com/bhojpur/service/pkg/secretstores/local/file"

	secretstores_loader "github.com/bhojpur/application/pkg/components/secretstores"

	// State Stores.
	"github.com/bhojpur/service/pkg/state"
	"github.com/bhojpur/service/pkg/state/aerospike"
	state_dynamodb "github.com/bhojpur/service/pkg/state/aws/dynamodb"
	state_azure_blobstorage "github.com/bhojpur/service/pkg/state/azure/blobstorage"
	state_cosmosdb "github.com/bhojpur/service/pkg/state/azure/cosmosdb"
	state_azure_tablestorage "github.com/bhojpur/service/pkg/state/azure/tablestorage"
	"github.com/bhojpur/service/pkg/state/cassandra"
	"github.com/bhojpur/service/pkg/state/couchbase"
	"github.com/bhojpur/service/pkg/state/gcp/firestore"
	"github.com/bhojpur/service/pkg/state/hashicorp/consul"
	"github.com/bhojpur/service/pkg/state/hazelcast"
	state_jetstream "github.com/bhojpur/service/pkg/state/jetstream"
	"github.com/bhojpur/service/pkg/state/memcached"
	"github.com/bhojpur/service/pkg/state/mongodb"
	state_mysql "github.com/bhojpur/service/pkg/state/mysql"
	state_oci_objectstorage "github.com/bhojpur/service/pkg/state/oci/objectstorage"
	state_oracledatabase "github.com/bhojpur/service/pkg/state/oracledatabase"
	"github.com/bhojpur/service/pkg/state/postgresql"
	state_redis "github.com/bhojpur/service/pkg/state/redis"
	"github.com/bhojpur/service/pkg/state/rethinkdb"
	"github.com/bhojpur/service/pkg/state/sqlserver"
	"github.com/bhojpur/service/pkg/state/zookeeper"

	state_loader "github.com/bhojpur/application/pkg/components/state"

	// Pub/Sub.
	configuration_loader "github.com/bhojpur/application/pkg/components/configuration"
	pubsub_loader "github.com/bhojpur/application/pkg/components/pubsub"
	pubs "github.com/bhojpur/service/pkg/pubsub"
	pubsub_snssqs "github.com/bhojpur/service/pkg/pubsub/aws/snssqs"

	// pubsub_eventhubs "github.com/bhojpur/service/pkg/pubsub/azure/eventhubs"
	// "github.com/bhojpur/service/pkg/pubsub/azure/servicebus"
	pubsub_gcp "github.com/bhojpur/service/pkg/pubsub/gcp/pubsub"
	pubsub_hazelcast "github.com/bhojpur/service/pkg/pubsub/hazelcast"
	pubsub_inmemory "github.com/bhojpur/service/pkg/pubsub/in-memory"
	pubsub_jetstream "github.com/bhojpur/service/pkg/pubsub/jetstream"
	pubsub_kafka "github.com/bhojpur/service/pkg/pubsub/kafka"
	pubsub_mqtt "github.com/bhojpur/service/pkg/pubsub/mqtt"
	"github.com/bhojpur/service/pkg/pubsub/natsstreaming"
	pubsub_pulsar "github.com/bhojpur/service/pkg/pubsub/pulsar"
	"github.com/bhojpur/service/pkg/pubsub/rabbitmq"
	pubsub_redis "github.com/bhojpur/service/pkg/pubsub/redis"

	// Name resolutions.
	nr "github.com/bhojpur/service/pkg/nameresolution"
	nr_consul "github.com/bhojpur/service/pkg/nameresolution/consul"
	nr_kubernetes "github.com/bhojpur/service/pkg/nameresolution/kubernetes"
	nr_mdns "github.com/bhojpur/service/pkg/nameresolution/mdns"

	nr_loader "github.com/bhojpur/application/pkg/components/nameresolution"

	// Bindings.
	"github.com/bhojpur/service/pkg/bindings"
	dingtalk_webhook "github.com/bhojpur/service/pkg/bindings/alicloud/dingtalk/webhook"
	"github.com/bhojpur/service/pkg/bindings/alicloud/oss"
	"github.com/bhojpur/service/pkg/bindings/alicloud/tablestore"
	"github.com/bhojpur/service/pkg/bindings/apns"
	"github.com/bhojpur/service/pkg/bindings/aws/dynamodb"
	"github.com/bhojpur/service/pkg/bindings/aws/kinesis"
	"github.com/bhojpur/service/pkg/bindings/aws/s3"
	"github.com/bhojpur/service/pkg/bindings/aws/ses"
	"github.com/bhojpur/service/pkg/bindings/aws/sns"
	"github.com/bhojpur/service/pkg/bindings/aws/sqs"
	"github.com/bhojpur/service/pkg/bindings/azure/blobstorage"
	bindings_cosmosdb "github.com/bhojpur/service/pkg/bindings/azure/cosmosdb"
	bindings_cosmosdbgremlinapi "github.com/bhojpur/service/pkg/bindings/azure/cosmosdbgremlinapi"
	"github.com/bhojpur/service/pkg/bindings/azure/eventgrid"

	// "github.com/bhojpur/service/pkg/bindings/azure/eventhubs"
	// "github.com/bhojpur/service/pkg/bindings/azure/servicebusqueues"
	"github.com/bhojpur/service/pkg/bindings/azure/signalr"
	"github.com/bhojpur/service/pkg/bindings/azure/storagequeues"
	"github.com/bhojpur/service/pkg/bindings/cron"
	"github.com/bhojpur/service/pkg/bindings/gcp/bucket"
	"github.com/bhojpur/service/pkg/bindings/gcp/pubsub"
	"github.com/bhojpur/service/pkg/bindings/graphql"
	"github.com/bhojpur/service/pkg/bindings/http"
	"github.com/bhojpur/service/pkg/bindings/influx"
	"github.com/bhojpur/service/pkg/bindings/kafka"
	"github.com/bhojpur/service/pkg/bindings/kubernetes"
	"github.com/bhojpur/service/pkg/bindings/localstorage"
	"github.com/bhojpur/service/pkg/bindings/mqtt"
	"github.com/bhojpur/service/pkg/bindings/mysql"
	"github.com/bhojpur/service/pkg/bindings/postgres"
	"github.com/bhojpur/service/pkg/bindings/postmark"
	bindings_rabbitmq "github.com/bhojpur/service/pkg/bindings/rabbitmq"
	"github.com/bhojpur/service/pkg/bindings/redis"
	"github.com/bhojpur/service/pkg/bindings/rethinkdb/statechange"
	"github.com/bhojpur/service/pkg/bindings/smtp"
	"github.com/bhojpur/service/pkg/bindings/twilio/sendgrid"
	"github.com/bhojpur/service/pkg/bindings/twilio/sms"
	"github.com/bhojpur/service/pkg/bindings/twitter"
	bindings_zeebe_command "github.com/bhojpur/service/pkg/bindings/zeebe/command"
	bindings_zeebe_jobworker "github.com/bhojpur/service/pkg/bindings/zeebe/jobworker"

	bindings_loader "github.com/bhojpur/application/pkg/components/bindings"

	// HTTP Middleware.

	middleware "github.com/bhojpur/service/pkg/middleware"
	"github.com/bhojpur/service/pkg/middleware/http/bearer"
	"github.com/bhojpur/service/pkg/middleware/http/oauth2"
	"github.com/bhojpur/service/pkg/middleware/http/oauth2clientcredentials"
	"github.com/bhojpur/service/pkg/middleware/http/opa"
	"github.com/bhojpur/service/pkg/middleware/http/ratelimit"
	"github.com/bhojpur/service/pkg/middleware/http/sentinel"

	http_middleware_loader "github.com/bhojpur/application/pkg/components/middleware/http"
	http_middleware "github.com/bhojpur/application/pkg/middleware/http"

	"github.com/bhojpur/service/pkg/configuration"
	configuration_redis "github.com/bhojpur/service/pkg/configuration/redis"
)

var (
	log        = logger.NewLogger("app.runtime")
	logService = logger.NewLogger("app.service")
)

func main() {
	// set GOMAXPROCS
	_, _ = maxprocs.Set()

	rt, err := runtime.FromFlags()
	if err != nil {
		log.Fatal(err)
	}

	err = rt.Run(
		runtime.WithSecretStores(
			secretstores_loader.New("kubernetes", func() secretstores.SecretStore {
				return secretstore_kubernetes.NewKubernetesSecretStore(logService)
			}),
			secretstores_loader.New("azure.keyvault", func() secretstores.SecretStore {
				return keyvault.NewAzureKeyvaultSecretStore(logService)
			}),
			secretstores_loader.New("hashicorp.vault", func() secretstores.SecretStore {
				return vault.NewHashiCorpVaultSecretStore(logService)
			}),
			secretstores_loader.New("aws.secretmanager", func() secretstores.SecretStore {
				return secretmanager.NewSecretManager(logService)
			}),
			secretstores_loader.New("aws.parameterstore", func() secretstores.SecretStore {
				return parameterstore.NewParameterStore(logService)
			}),
			secretstores_loader.New("gcp.secretmanager", func() secretstores.SecretStore {
				return gcp_secretmanager.NewSecreteManager(logService)
			}),
			secretstores_loader.New("local.file", func() secretstores.SecretStore {
				return secretstore_file.NewLocalSecretStore(logService)
			}),
			secretstores_loader.New("local.env", func() secretstores.SecretStore {
				return secretstore_env.NewEnvSecretStore(logService)
			}),
			secretstores_loader.New("alicloud.parameterstore", func() secretstores.SecretStore {
				return alicloud_paramstore.NewParameterStore(logService)
			}),
		),
		runtime.WithStates(
			state_loader.New("redis", func() state.Store {
				return state_redis.NewRedisStateStore(logService)
			}),
			state_loader.New("consul", func() state.Store {
				return consul.NewConsulStateStore(logService)
			}),
			state_loader.New("azure.blobstorage", func() state.Store {
				return state_azure_blobstorage.NewAzureBlobStorageStore(logService)
			}),
			state_loader.New("azure.cosmosdb", func() state.Store {
				return state_cosmosdb.NewCosmosDBStateStore(logService)
			}),
			state_loader.New("azure.tablestorage", func() state.Store {
				return state_azure_tablestorage.NewAzureTablesStateStore(logService)
			}),
			state_loader.New("cassandra", func() state.Store {
				return cassandra.NewCassandraStateStore(logService)
			}),
			state_loader.New("memcached", func() state.Store {
				return memcached.NewMemCacheStateStore(logService)
			}),
			state_loader.New("mongodb", func() state.Store {
				return mongodb.NewMongoDB(logService)
			}),
			state_loader.New("zookeeper", func() state.Store {
				return zookeeper.NewZookeeperStateStore(logService)
			}),
			state_loader.New("gcp.firestore", func() state.Store {
				return firestore.NewFirestoreStateStore(logService)
			}),
			state_loader.New("postgresql", func() state.Store {
				return postgresql.NewPostgreSQLStateStore(logService)
			}),
			state_loader.New("sqlserver", func() state.Store {
				return sqlserver.NewSQLServerStateStore(logService)
			}),
			state_loader.New("hazelcast", func() state.Store {
				return hazelcast.NewHazelcastStore(logService)
			}),
			state_loader.New("couchbase", func() state.Store {
				return couchbase.NewCouchbaseStateStore(logService)
			}),
			state_loader.New("aerospike", func() state.Store {
				return aerospike.NewAerospikeStateStore(logService)
			}),
			state_loader.New("rethinkdb", func() state.Store {
				return rethinkdb.NewRethinkDBStateStore(logService)
			}),
			state_loader.New("aws.dynamodb", state_dynamodb.NewDynamoDBStateStore),
			state_loader.New("mysql", func() state.Store {
				return state_mysql.NewMySQLStateStore(logService)
			}),
			state_loader.New("oci.objectstorage", func() state.Store {
				return state_oci_objectstorage.NewOCIObjectStorageStore(logService)
			}),
			state_loader.New("jetstream", func() state.Store {
				return state_jetstream.NewJetstreamStateStore(logService)
			}),
			state_loader.New("oracledatabase", func() state.Store {
				return state_oracledatabase.NewOracleDatabaseStateStore(logService)
			}),
		),
		runtime.WithConfigurations(
			configuration_loader.New("redis", func() configuration.Store {
				return configuration_redis.NewRedisConfigurationStore(logService)
			}),
		),
		runtime.WithPubSubs(
			/*
				pubsub_loader.New("azure.eventhubs", func() pubs.PubSub {
					return pubsub_eventhubs.NewAzureEventHubs(logService)
				}),

					pubsub_loader.New("azure.servicebus", func() pubs.PubSub {
						return servicebus.NewAzureServiceBus(logService)
					}),*/
			pubsub_loader.New("gcp.pubsub", func() pubs.PubSub {
				return pubsub_gcp.NewGCPPubSub(logService)
			}),
			pubsub_loader.New("hazelcast", func() pubs.PubSub {
				return pubsub_hazelcast.NewHazelcastPubSub(logService)
			}),
			pubsub_loader.New("jetstream", func() pubs.PubSub {
				return pubsub_jetstream.NewJetStream(logService)
			}),
			pubsub_loader.New("kafka", func() pubs.PubSub {
				return pubsub_kafka.NewKafka(logService)
			}),
			pubsub_loader.New("mqtt", func() pubs.PubSub {
				return pubsub_mqtt.NewMQTTPubSub(logService)
			}),
			pubsub_loader.New("natsstreaming", func() pubs.PubSub {
				return natsstreaming.NewNATSStreamingPubSub(logService)
			}),
			pubsub_loader.New("pulsar", func() pubs.PubSub {
				return pubsub_pulsar.NewPulsar(logService)
			}),
			pubsub_loader.New("rabbitmq", func() pubs.PubSub {
				return rabbitmq.NewRabbitMQ(logService)
			}),
			pubsub_loader.New("redis", func() pubs.PubSub {
				return pubsub_redis.NewRedisStreams(logService)
			}),
			pubsub_loader.New("snssqs", func() pubs.PubSub {
				return pubsub_snssqs.NewSnsSqs(logService)
			}),
			pubsub_loader.New("in-memory", func() pubs.PubSub {
				return pubsub_inmemory.New(logService)
			}),
		),
		runtime.WithNameResolutions(
			nr_loader.New("mdns", func() nr.Resolver {
				return nr_mdns.NewResolver(logService)
			}),
			nr_loader.New("kubernetes", func() nr.Resolver {
				return nr_kubernetes.NewResolver(logService)
			}),
			nr_loader.New("consul", func() nr.Resolver {
				return nr_consul.NewResolver(logService)
			}),
		),
		runtime.WithInputBindings(
			bindings_loader.NewInput("aws.sqs", func() bindings.InputBinding {
				return sqs.NewAWSSQS(logService)
			}),
			bindings_loader.NewInput("aws.kinesis", func() bindings.InputBinding {
				return kinesis.NewAWSKinesis(logService)
			}),
			bindings_loader.NewInput("azure.eventgrid", func() bindings.InputBinding {
				return eventgrid.NewAzureEventGrid(logService)
			}),
			/*
				bindings_loader.NewInput("azure.eventhubs", func() bindings.InputBinding {
					return eventhubs.NewAzureEventHubs(logService)
				}),

					bindings_loader.NewInput("azure.servicebusqueues", func() bindings.InputBinding {
						return servicebusqueues.NewAzureServiceBusQueues(logService)
					}), */
			bindings_loader.NewInput("azure.storagequeues", func() bindings.InputBinding {
				return storagequeues.NewAzureStorageQueues(logService)
			}),
			bindings_loader.NewInput("cron", func() bindings.InputBinding {
				return cron.NewCron(logService)
			}),
			bindings_loader.NewInput("dingtalk.webhook", func() bindings.InputBinding {
				return dingtalk_webhook.NewDingTalkWebhook(logService)
			}),
			bindings_loader.NewInput("gcp.pubsub", func() bindings.InputBinding {
				return pubsub.NewGCPPubSub(logService)
			}),
			bindings_loader.NewInput("kafka", func() bindings.InputBinding {
				return kafka.NewKafka(logService)
			}),
			bindings_loader.NewInput("kubernetes", func() bindings.InputBinding {
				return kubernetes.NewKubernetes(logService)
			}),
			bindings_loader.NewInput("mqtt", func() bindings.InputBinding {
				return mqtt.NewMQTT(logService)
			}),
			bindings_loader.NewInput("rabbitmq", func() bindings.InputBinding {
				return bindings_rabbitmq.NewRabbitMQ(logService)
			}),
			bindings_loader.NewInput("rethinkdb.statechange", func() bindings.InputBinding {
				return statechange.NewRethinkDBStateChangeBinding(logService)
			}),
			bindings_loader.NewInput("twitter", func() bindings.InputBinding {
				return twitter.NewTwitter(logService)
			}),
			bindings_loader.NewInput("zeebe.jobworker", func() bindings.InputBinding {
				return bindings_zeebe_jobworker.NewZeebeJobWorker(logService)
			}),
		),
		runtime.WithOutputBindings(
			bindings_loader.NewOutput("alicloud.oss", func() bindings.OutputBinding {
				return oss.NewAliCloudOSS(logService)
			}),
			bindings_loader.NewOutput("alicloud.tablestore", func() bindings.OutputBinding {
				return tablestore.NewAliCloudTableStore(log)
			}),
			bindings_loader.NewOutput("apns", func() bindings.OutputBinding {
				return apns.NewAPNS(logService)
			}),
			bindings_loader.NewOutput("aws.s3", func() bindings.OutputBinding {
				return s3.NewAWSS3(logService)
			}),
			bindings_loader.NewOutput("aws.ses", func() bindings.OutputBinding {
				return ses.NewAWSSES(logService)
			}),
			bindings_loader.NewOutput("aws.sqs", func() bindings.OutputBinding {
				return sqs.NewAWSSQS(logService)
			}),
			bindings_loader.NewOutput("aws.sns", func() bindings.OutputBinding {
				return sns.NewAWSSNS(logService)
			}),
			bindings_loader.NewOutput("aws.kinesis", func() bindings.OutputBinding {
				return kinesis.NewAWSKinesis(logService)
			}),
			bindings_loader.NewOutput("aws.dynamodb", func() bindings.OutputBinding {
				return dynamodb.NewDynamoDB(logService)
			}),
			bindings_loader.NewOutput("azure.blobstorage", func() bindings.OutputBinding {
				return blobstorage.NewAzureBlobStorage(logService)
			}),
			bindings_loader.NewOutput("azure.cosmosdb", func() bindings.OutputBinding {
				return bindings_cosmosdb.NewCosmosDB(logService)
			}),
			bindings_loader.NewOutput("azure.cosmosdb.gremlinapi", func() bindings.OutputBinding {
				return bindings_cosmosdbgremlinapi.NewCosmosDBGremlinAPI(logService)
			}),
			bindings_loader.NewOutput("azure.eventgrid", func() bindings.OutputBinding {
				return eventgrid.NewAzureEventGrid(logService)
			}),
			/*
				bindings_loader.NewOutput("azure.eventhubs", func() bindings.OutputBinding {
					return eventhubs.NewAzureEventHubs(logService)
				}),

					bindings_loader.NewOutput("azure.servicebusqueues", func() bindings.OutputBinding {
						return servicebusqueues.NewAzureServiceBusQueues(logService)
					}), */
			bindings_loader.NewOutput("azure.signalr", func() bindings.OutputBinding {
				return signalr.NewSignalR(logService)
			}),
			bindings_loader.NewOutput("azure.storagequeues", func() bindings.OutputBinding {
				return storagequeues.NewAzureStorageQueues(logService)
			}),
			bindings_loader.NewOutput("cron", func() bindings.OutputBinding {
				return cron.NewCron(logService)
			}),
			bindings_loader.NewOutput("dingtalk.webhook", func() bindings.OutputBinding {
				return dingtalk_webhook.NewDingTalkWebhook(logService)
			}),
			bindings_loader.NewOutput("gcp.bucket", func() bindings.OutputBinding {
				return bucket.NewGCPStorage(logService)
			}),
			bindings_loader.NewOutput("gcp.pubsub", func() bindings.OutputBinding {
				return pubsub.NewGCPPubSub(logService)
			}),
			bindings_loader.NewOutput("http", func() bindings.OutputBinding {
				return http.NewHTTP(logService)
			}),
			bindings_loader.NewOutput("influx", func() bindings.OutputBinding {
				return influx.NewInflux(logService)
			}),
			bindings_loader.NewOutput("kafka", func() bindings.OutputBinding {
				return kafka.NewKafka(logService)
			}),
			bindings_loader.NewOutput("localstorage", func() bindings.OutputBinding {
				return localstorage.NewLocalStorage(logService)
			}),
			bindings_loader.NewOutput("mqtt", func() bindings.OutputBinding {
				return mqtt.NewMQTT(logService)
			}),
			bindings_loader.NewOutput("mysql", func() bindings.OutputBinding {
				return mysql.NewMysql(logService)
			}),
			bindings_loader.NewOutput("postgres", func() bindings.OutputBinding {
				return postgres.NewPostgres(logService)
			}),
			bindings_loader.NewOutput("postmark", func() bindings.OutputBinding {
				return postmark.NewPostmark(logService)
			}),
			bindings_loader.NewOutput("rabbitmq", func() bindings.OutputBinding {
				return bindings_rabbitmq.NewRabbitMQ(logService)
			}),
			bindings_loader.NewOutput("redis", func() bindings.OutputBinding {
				return redis.NewRedis(logService)
			}),
			bindings_loader.NewOutput("smtp", func() bindings.OutputBinding {
				return smtp.NewSMTP(logService)
			}),
			bindings_loader.NewOutput("twilio.sms", func() bindings.OutputBinding {
				return sms.NewSMS(logService)
			}),
			bindings_loader.NewOutput("twilio.sendgrid", func() bindings.OutputBinding {
				return sendgrid.NewSendGrid(logService)
			}),
			bindings_loader.NewOutput("twitter", func() bindings.OutputBinding {
				return twitter.NewTwitter(logService)
			}),
			bindings_loader.NewOutput("zeebe.command", func() bindings.OutputBinding {
				return bindings_zeebe_command.NewZeebeCommand(logService)
			}),
			bindings_loader.NewOutput("graphql", func() bindings.OutputBinding {
				return graphql.NewGraphQL(logService)
			}),
		),
		runtime.WithHTTPMiddleware(
			http_middleware_loader.New("uppercase", func(metadata middleware.Metadata) (http_middleware.Middleware, error) {
				return func(h fasthttp.RequestHandler) fasthttp.RequestHandler {
					return func(ctx *fasthttp.RequestCtx) {
						body := string(ctx.PostBody())
						ctx.Request.SetBody([]byte(strings.ToUpper(body)))
						h(ctx)
					}
				}, nil
			}),
			http_middleware_loader.New("oauth2", func(metadata middleware.Metadata) (http_middleware.Middleware, error) {
				return oauth2.NewOAuth2Middleware().GetHandler(metadata)
			}),
			http_middleware_loader.New("oauth2clientcredentials", func(metadata middleware.Metadata) (http_middleware.Middleware, error) {
				return oauth2clientcredentials.NewOAuth2ClientCredentialsMiddleware(log).GetHandler(metadata)
			}),
			http_middleware_loader.New("ratelimit", func(metadata middleware.Metadata) (http_middleware.Middleware, error) {
				return ratelimit.NewRateLimitMiddleware(log).GetHandler(metadata)
			}),
			http_middleware_loader.New("bearer", func(metadata middleware.Metadata) (http_middleware.Middleware, error) {
				return bearer.NewBearerMiddleware(log).GetHandler(metadata)
			}),
			http_middleware_loader.New("opa", func(metadata middleware.Metadata) (http_middleware.Middleware, error) {
				return opa.NewMiddleware(log).GetHandler(metadata)
			}),
			http_middleware_loader.New("sentinel", func(metadata middleware.Metadata) (http_middleware.Middleware, error) {
				return sentinel.NewMiddleware(log).GetHandler(metadata)
			}),
		),
	)
	if err != nil {
		log.Fatalf("fatal error from Bhojpur Application runtime: %s", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	<-stop
	rt.ShutdownWithWait()
}
