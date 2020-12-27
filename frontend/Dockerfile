# base image
FROM node:15.5.0-alpine as build

# set working directory
WORKDIR /mainpage

# add `/app/node_modules/.bin` to $PATH
ENV PATH /mainpage/node_modules/.bin:$PATH

# install and cache app dependencies
RUN yarn global add react-scripts@3.4.0

COPY ./package.json ./package.json
RUN yarn install
COPY ./  ./
RUN yarn run build

COPY start.sh /opt/start.sh
RUN chmod +x /opt/start.sh

# production environment
FROM nginx:alpine
COPY --from=build /mainpage/build /usr/share/nginx/html
COPY --from=build /opt/start.sh /opt/start.sh
EXPOSE 80

# start app
CMD ["/opt/start.sh"]