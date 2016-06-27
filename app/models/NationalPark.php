<?php

class NationalPark extends \Eloquent {

	protected $table = 'national_parks';

	// Add your validation rules here
	public static $rules = [
		'name'    			=> 'required|min:1|max:240',
	    'location'     		=> 'required|min:1|max:240',
	    'date_established'  => 'required|date',
	    'area_in_acres'	    => 'required|min:1|max:99999999999',
	    'description'		=> 'required|min:1|max:240'
	];

	// Don't forget to fill this array
	protected $fillable = array('name', 'location', 'date_established', 'area_in_acres', 'description');

}