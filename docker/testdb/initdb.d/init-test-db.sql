use mysql;

CREATE USER 'test'@'%' IDENTIFIED BY 'test';

GRANT
    ALL PRIViLEGES
    ON *.*
    TO 'test'@'%'
;

CREATE DATABASE `test_db`;
