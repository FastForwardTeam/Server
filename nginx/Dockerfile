ARG exposedport=80
# Pull nginx base image
FROM nginx:latest

# Expost port 80
EXPOSE $exposedport

# Copy custom configuration file from the current directory
COPY nginx.conf /etc/nginx/nginx.conf
