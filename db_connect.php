<?php
// Get new instance of PDO object
$dbc = new PDO('mysql:host='. $_ENV['DB_HOST'] .';dbname='. $_ENV['DB_NAME'], $_ENV['DB_USER'], $_ENV['DB_PASS']);

// Tell PDO to throw exceptions on error
$dbc->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);


