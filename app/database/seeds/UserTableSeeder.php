<?php

class UserTableSeeder extends Seeder {

	public function run()
	{
		$user = new User();
		$user->first_name = 'Pascal';
		$user->last_name = 'Allen';
		$user->username = 'pascalallen';
		$user->email = 'thomaspascalallen@yahoo.com';
		$user->password = $_ENV['USER_PASS'];
		$user->birthday = '1988-05-13';
		$user->phone_number = '5555555555';
		$user->zipcode = '78108';
		$user->save();
	}
}