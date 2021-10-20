# Deploying

1) Setup mariadb repo by running [this script](https://mariadb.com/kb/en/mariadb-package-repository-setup-and-usage/).

2) Install it 

`sudo apt install mariadb-server`

3) Secure the install by:

`sudo mysql_secure_installation` 

4) Run `mariadb -u root -p` and then setup the db by-

Here, we're using ffdb as the name feel free to change it

I) `CREATE DATABASE ffdb;` <br>
II) `USE ffdb`  <br>
Now run the three commands listed in [schema.sql](./schema.sql)

5) Download the latest setup from the releases page and run it with sudo 
