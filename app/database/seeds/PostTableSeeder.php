<?php

class PostTableSeeder extends Seeder {

	public function run()
	{
		$post1 = new Post();
		$post1->title = 'this is a test title';
		$post1->body = 'this is a test body for poststableseeder';
		$post1->user_id = '1';
		$post1->save();

		$post2 = new Post();
		$post2->title = 'this is a test title2';
		$post2->body = 'this is a test body for poststableseeder2';
		$post2->user_id = '1';
		$post2->save();

		$post3 = new Post();
		$post3->title = 'this is a test title3';
		$post3->body = 'this is a test body for poststableseeder3';
		$post3->user_id = '1';
		$post3->save();

		$post4 = new Post();
		$post4->title = 'this is a test title4';
		$post4->body = 'this is a test body for poststableseeder4';
		$post4->user_id = '1';
		$post4->save();

		$post5 = new Post();
		$post5->title = 'this is a test title5';
		$post5->body = 'this is a test body for poststableseeder5';
		$post5->user_id = '1';
		$post5->save();
	}
}