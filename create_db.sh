echo "drop database kerbspace" | mysql -uroot
echo "create database kerbspace" | mysql -uroot
cat db_schema.sql | mysql -uroot kerbspace
