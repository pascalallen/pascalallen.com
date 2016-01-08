<?php

use Illuminate\Database\Schema\Blueprint;
use Illuminate\Database\Migrations\Migration;

class CreateUsersTableMigration extends Migration {

	/**
	 * Run the migrations.
	 *
	 * @return void
	 */
	public function up()
	{
		Schema::create('users', function(Blueprint $table)
		{
			$table->increments('id');
			$table->string('first_name', 100);
			$table->string('last_name', 100);
			$table->string('username', 100)->unique();
			$table->string('email', 100)->unique();
			$table->string('password', 255);
			$table->date('birthday');
			$table->string('phone_number', 100)->nullable();
			$table->integer('zipcode');
			$table->timestamps();
		});
	}

	/**
	 * Reverse the migrations.
	 *
	 * @return void
	 */
	public function down()
	{
		Schema::drop('users');
	}

}
