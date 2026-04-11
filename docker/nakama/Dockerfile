FROM registry.heroiclabs.com/heroiclabs/nakama-pluginbuilder:3.38.0 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build --trimpath --buildmode=plugin -o /workspace/build/lila.so ./nakama

FROM registry.heroiclabs.com/heroiclabs/nakama:3.38.0

COPY --from=builder /workspace/build/lila.so /nakama/data/modules/lila.so
