FROM golang:1.27.3-trixie AS deps

WORKDIR /src
ENV CGO_ENABLED=0
COPY go.mod go.sum ./

FROM deps AS api-build
COPY cmd/api ./cmd/api
COPY internal ./internal
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -o /out/api ./cmd/api

FROM deps AS goose-build
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    goose_version=$(go list -m -f '{{.Version}}' github.com/pressly/goose/v3) && \
    GOBIN=/out go install -trimpath \
      -tags='no_clickhouse no_mssql no_mysql no_sqlite3 no_libsql no_vertica no_ydb' \
      github.com/pressly/goose/v3/cmd/goose@${goose_version}

FROM scratch AS api
WORKDIR /
USER 65532:65532
COPY --from=api-build /out/api /api
COPY env.example.yml /env.local.yml
ENTRYPOINT ["/api"]

FROM scratch AS migrate
USER 65532:65532
COPY --from=goose-build /out/goose /goose
COPY migrations/*.sql /migrations/
ENTRYPOINT ["/goose"]
CMD ["-env=none", "-dir=/migrations", "-timeout=1m", "up"]