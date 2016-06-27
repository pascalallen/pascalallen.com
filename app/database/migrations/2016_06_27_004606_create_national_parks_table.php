<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;

class CreateNationalParksTable extends Migration {

	/**
	 * Run the migrations.
	 *
	 * @return void
	 */
	public function up()
	{
		Schema::create('national_parks', function(Blueprint $table)
		{
			$table->increments('id');
			$table->string('name', 240);
			$table->string('location', 50);
			$table->date('date_established');
			$table->double('area_in_acres', 10, 2);
			$table->text('description');
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
		Schema::drop('national_parks');
	}

}
