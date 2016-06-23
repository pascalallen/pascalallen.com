<?php
require_once 'parks_logins.php';

require 'db_connect.php';

$truncate_all = 'TRUNCATE national_parks';

$dbc->exec($truncate_all);

$nationalParks = [
    ['name' => 'Acadia',           'location' => '44.35°N 68.21°W',  'date_established' => '1919-02-26', 'area_in_acres' => 47389.67, 'description'  => 'Covering most of Mount Desert Island and other coastal islands, Acadia features the tallest mountain on the Atlantic coast of the United States, granite peaks, ocean shoreline, woodlands, and lakes. There are freshwater, estuary, forest, and intertidal habitats.'],
    ['name' => 'American Samoa',   'location' => '14.25°S 170.68°W', 'date_established' => '1988-10-31', 'area_in_acres' => 9000.00, 'description'   => 'The southernmost national park is on three Samoan islands and protects coral reefs, rainforests, volcanic mountains, and white beaches. The area is also home to flying foxes, brown boobies, sea turtles, and 900 species of fish.'],
    ['name' => 'Arches',           'location' => '38.68°N 109.57°W', 'date_established' => '1929-04-12', 'area_in_acres' => 76518.98, 'description'  => 'This site features more than 2,000 natural sandstone arches, including the famous Delicate Arch. In a desert climate, millions of years of erosion have led to these structures, and the arid ground has life-sustaining soil crust and potholes, which serve as natural water-collecting basins. Other geologic formations are stone columns, spires, fins, and towers.'],
    ['name' => 'Badlands',         'location' => '43.75°N 102.50°W', 'date_established' => '1978-11-10', 'area_in_acres' => 242755.94, 'description' => 'The Badlands are a collection of buttes, pinnacles, spires, and grass prairies. It has the world\'s richest fossil beds from the Oligocene epoch, and the wildlife includes bison, bighorn sheep, black-footed ferrets, and swift foxes.'],
    ['name' => 'Big Bend',         'location' => '29.25°N 103.25°W', 'date_established' => '1944-06-12', 'area_in_acres' => 801163.21, 'description' => 'Named for the prominent bend in the Rio Grande along the US–Mexico border, this park encompasses a large and remote part of the Chihuahuan Desert. Its main attraction is backcountry recreation in the arid Chisos Mountains and in canyons along the river. A wide variety of Cretaceous and Tertiary fossils as well as cultural artifacts of Native Americans also exist within its borders.'],
    ['name' => 'Biscayne',         'location' => '25.65°N 80.08°W',  'date_established' => '1980-06-28', 'area_in_acres' => 172924.07, 'description' => 'Located in Biscayne Bay, this park at the north end of the Florida Keys has four interrelated marine ecosystems: mangrove forest, the Bay, the Keys, and coral reefs. Threatened animals include the West Indian manatee, American crocodile, various sea turtles, and peregrine falcon.'],
    ['name' => 'Bryce Canyon',     'location' => '37.57°N 112.18°W', 'date_established' => '1928-02-25', 'area_in_acres' => 35835.08, 'description'  => 'Bryce Canyon is a giant geological amphitheater on the Paunsaugunt Plateau. The unique area has hundreds of tall sandstone hoodoos formed by erosion. The region was originally settled by Native Americans and later by Mormon pioneers.'],
    ['name' => 'Canyonlands',      'location' => '38.2°N 109.93°W',  'date_established' => '1964-09-12', 'area_in_acres' => 337597.83, 'description' => 'This landscape was eroded into a maze of canyons, buttes, and mesas by the combined efforts of the Colorado River, Green River, and their tributaries, which divide the park into three districts. There are rock pinnacles and other naturally sculpted rock formations, as well as artifacts from Ancient Pueblo peoples.'],
    ['name' => 'Capitol Reef', 	   'location' => '38.20°N 111.17°W', 'date_established' => '1971-12-18', 'area_in_acres' => 241904.26, 'description' => 'The park\'s Waterpocket Fold is a 100-mile (160 km) monocline that exhibits the earth\'s diverse geologic layers. Other natural features are monoliths, sandstone domes, and cliffs shaped like the United States Capitol.'],
    ['name' => 'Carlsbad Caverns', 'location' => '32.17°N 104.44°W', 'date_established' => '1930-05-14', 'area_in_acres' => 46766.45, 'description'  => 'Carlsbad Caverns has 117 caves, the longest of which is over 120 miles (190 km) long. The Big Room is almost 4,000 feet (1,200 m) long, and the caves are home to over 400,000 Mexican free-tailed bats and sixteen other species. Above ground are the Chihuahuan Desert and Rattlesnake Springs.']
];

foreach ($nationalParks as $park) {
    $create_table = "INSERT INTO national_parks (name, location, date_established, area_in_acres, description) VALUES (:name, :location, :date_established, :area_in_acres, :description)";

    $stmt = $dbc->prepare($create_table);
    $stmt->bindValue(':name', $park['name'], PDO::PARAM_STR);
    $stmt->bindValue(':location', $park['location'], PDO::PARAM_STR);
    $stmt->bindValue(':date_established', $park['date_established'], PDO::PARAM_STR);
    $stmt->bindValue(':area_in_acres', $park['area_in_acres'], PDO::PARAM_STR);
    $stmt->bindValue(':description', $park['description'], PDO::PARAM_STR);

    $stmt->execute();
}
