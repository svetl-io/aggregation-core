ARG BUILDER_IMAGE=golang:1.22-alpine
ARG DISTROLESS_IMAGE=alpine

############################
# STEP 1 build executable binary
############################
FROM ${BUILDER_IMAGE} as builder
WORKDIR /app
RUN apk update && apk add build-base
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
# RUN go test --cover ./... -v -coverprofile=coverage.out
# RUN go tool cover -func=coverage.out
RUN apk add alpine-sdk 
RUN GOOS=linux GOARCH=amd64 go build -tags musl -buildvcs=false -o bin/aggregation-core .

# ############################
# # STEP 2 build a small image
# ############################
FROM ${DISTROLESS_IMAGE}
# Copy our static executable
COPY --from=builder /app/bin/aggregation-core ./aggregation-core
ENTRYPOINT ["/aggregation-core"]