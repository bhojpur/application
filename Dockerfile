FROM moby/buildkit:v0.9.3
WORKDIR /application
COPY application README.md /application/
ENV PATH=/application:$PATH
ENTRYPOINT [ "/bhojpur/application" ]