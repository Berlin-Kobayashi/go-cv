FROM node:8.11 as build

WORKDIR /usr/src/app/

COPY . .

RUN npm install

FROM nginx:1.15.1

COPY --from=build /usr/src/app /usr/share/nginx/html/
