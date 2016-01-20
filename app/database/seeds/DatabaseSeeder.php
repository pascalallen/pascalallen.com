<?php

class DatabaseSeeder extends Seeder {

	/**
	 * Run the database seeds.
	 *
	 * @return void
	 */
	public function run()
	{
		Eloquent::unguard();

		DB::table('posts')->delete();
		DB::table('users')->delete();
		
		// DELETE DATA OPPOSITE WAY THAT YOU ADD DATA
		$this->call('UserTableSeeder');
		$this->call('PostTableSeeder');
	}

}
