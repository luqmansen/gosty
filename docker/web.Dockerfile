FROM node:15.11.0-alpine3.10 as builder

ARG APISERVER_HOST=http://192.168.59.2:30800
ARG FILESERVER_HOST=http://192.168.59.2:30801

ENV REACT_APP_APISERVER_HOST=$APISERVER_HOST
ENV REACT_APP_FILESERVER_HOST=$FILESERVER_HOST

WORKDIR '/app'

COPY web/package*.json .
RUN npm install

COPY web/ .
RUN npm run build

FROM nginx:1.19.6-alpine
EXPOSE 80
COPY --from=builder /app/build /usr/share/nginx/html