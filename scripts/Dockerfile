FROM ubuntu as build
RUN apt-get update
RUN apt-get install -y curl git
RUN curl https://get.extism.org/cli | sh -s -- -y -o /
RUN /extism -h # smoke test the binary is installed here

FROM scratch
WORKDIR /
COPY --from=build /extism .
