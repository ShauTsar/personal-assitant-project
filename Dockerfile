# Build backend
FROM golang:1.22.2-alpine as backend
WORKDIR /app
COPY go.mod .
COPY go.sum .
COPY config .
COPY personal-assitant-server .
RUN go mod tidy
RUN go build -o grpc-main ./grpc
RUN go build -o botapi-main ./botAPI

# Build frontend
FROM node:21-alpine as frontend
WORKDIR /app
COPY personal-assistant-web/package*.json ./
RUN npm install
COPY personal-assistant-web .
RUN npm run build

# Final image
FROM alpine:3.17
COPY --from=backend /app/grpc-main /app/grpc-main
COPY --from=backend /app/botapi-main /app/botapi-main
COPY --from=frontend /app/build /app/personal-assistant-web
ENTRYPOINT ["/app/grpc-main"]
CMD ["/app/botapi-main"]

# Postgres
FROM postgres:16.2-alpine
ENV POSTGRES_DB=novaDB
ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=NNA2s*123

# Redis
FROM redis:7.2.4-alpine

# Elasticsearch
FROM elasticsearch:8.13.0


## Backend (gRPC и botAPI)
#FROM golang:1.22.2-alpine as backend
#WORKDIR /app
#COPY personal-assitant-project/personal-assitant-server/grpc .
#RUN go build -o grpc-main .
#COPY personal-assitant-project/personal-assitant-server/botAPI .
#RUN go build -o botapi-main .
#
## Frontend (React)
#FROM node:21-alpine as frontend
#WORKDIR /app
#COPY personal-assitant-project/personal-assistant-web .
#RUN npm install
#RUN npm run build
#
## Nginx
#FROM nginx:stable-alpine
#COPY --from=frontend /app/build /usr/share/nginx/html
#EXPOSE 80
#CMD ["nginx", "-g", "daemon off;"]
#
## Postgres
#FROM postgres:16.2-alpine
#ENV POSTGRES_DB=novaDB
#ENV POSTGRES_USER=postgres
#ENV POSTGRES_PASSWORD=NNA2s*123
#
## Redis
#FROM redis:7.2.4-alpine
#
## Elasticsearch
#FROM elasticsearch:8.13.0
#
#


#FROM golang:1.22.2-alpine
#WORKDIR /app
#COPY personal-assitant-project/personal-assitant-server/grpc .
#RUN go build -o main .
#CMD ["./main"]
#
#FROM golang:1.22.2-alpine
#WORKDIR /app
#COPY personal-assitant-project/personal-assitant-server/botAPI .
#RUN go build -o main .
#
#CMD ["./main"]
#FROM node:21-alpine as build
#WORKDIR /app
#COPY personal-assitant-project/personal-assistant-web .
#RUN npm install
#RUN npm run build
#
#FROM nginx:stable-alpine
#COPY --from=build /app/build /usr/share/nginx/html
#EXPOSE 80
#CMD ["nginx", "-g", "daemon off;"]
#
#FROM postgres:16.2-alpine
#ENV POSTGRES_DB=novaDB
#ENV POSTGRES_USER=postgres
#ENV POSTGRES_PASSWORD=NNA2s*123
#
#FROM redis:7.2.4-alpine
#
#FROM elasticsearch:8.13.0