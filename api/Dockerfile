FROM golang:1.24 AS build
WORKDIR /src
COPY ./ ./
RUN go mod tidy && go mod vendor
RUN CGO_ENABLED=0 go build -o /src/api ./api.go


FROM scratch
COPY --from=build /src/api /bin/api

CMD ["/bin/api"]