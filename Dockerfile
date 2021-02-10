FROM golang:1.15 AS build
WORKDIR /src
COPY . .
RUN go build -o /app

FROM gcr.io/distroless/base
COPY --from=build /app /
CMD [ "/app" ]