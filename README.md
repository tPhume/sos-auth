# SOS Authentication
Gives out JWT and refresh token.

## Local testing server
cd into the root of this directory on your local machine. Run `docker-compose up` to initialize redis,postgresql (this comes with one data populated inside it) and the webserver. You can add more users on intializatoin through init.sql if you wish. If you change anything, run `docker-compose up --build` to rebuild the images.