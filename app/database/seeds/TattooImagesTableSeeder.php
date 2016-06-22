<?php

// Composer: "fzaninotto/faker": "v1.3.0"
use Faker\Factory as Faker;

class TattooImagesTableSeeder extends Seeder {

	public function run()
	{
		// $faker = Faker::create();

		$dir = "public/img/*.JPG";

		//returns array matching $dir
		$images = glob($dir);

		foreach($images as $image)
		{
			TattooImage::create([
				'title' => '',
				'body' 	=> '',
				'image' => $image
			]);
			var_dump($image);
		}
	}

}