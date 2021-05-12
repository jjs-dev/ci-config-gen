FROM golang:1.16.4 as build
COPY . /app
WORKDIR /app
RUN go build -v  -o /gen

FROM gcr.io/distroless/static
COPY --from=build /gen /ci-config-gen
ENTRYPOINT ["/ci-config-gen"]
CMD ["--repo-root", "/src"]
VOLUME ["/src"]