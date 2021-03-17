FROM node:15.11.0-alpine3.10 as builder
WORKDIR '/app'

COPY web/package*.json ./
RUN npm install
COPY web/ .
RUN npm run build

FROM nginx:1.19.6-alpine
EXPOSE 80
COPY --from=builder /app/build /usr/share/nginx/html