FROM golang:1.20-alpine as build

WORKDIR /src
COPY . ./
RUN go mod download
RUN go build -o /yogai ./service.go

FROM golang:1.20-alpine

COPY --from=build /yogai /yogai
CMD /yogai