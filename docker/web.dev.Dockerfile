FROM nginx:1.19.6-alpine

# Dockerfile for dev, run the build on local
ADD ./web/config/default.conf /etc/nginx/conf.d/default.conf
ADD ./web/config/nginx.conf /etc/nginx/nginx.conf
ADD ./web/build /usr/share/nginx/html


EXPOSE 80