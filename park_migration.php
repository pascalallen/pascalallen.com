<?php
require_once 'parks_logins.php';

require 'db_connect.php';

// echo $dbc->getAttribute(PDO::ATTR_CONNECTION_STATUS) . "\n";

$drop_table = "DROP TABLE IF EXISTS national_parks";

$dbc->exec($drop_table);

$create_table = 'CREATE TABLE national_parks (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(240) NOT NULL,
    location VARCHAR(50) NOT NULL,
    date_established DATE NOT NULL,
    area_in_acres DOUBLE(10, 2) NOT NULL,
    description TEXT NOT NULL,
    PRIMARY KEY (id)
)';

$dbc->exec($create_table);

require_once 'park_seeder.php';