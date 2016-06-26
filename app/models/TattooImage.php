<?php

class TattooImage extends \Eloquent {

	protected $table = 'tattoo_images';
	

	// Add your validation rules here
	public static $rules = [
		'title'    => 'max:100',
	    'body'     => 'max:500',
	    'image'	   => 'image'
	];

	// Don't forget to fill this array
	protected $fillable = array('title', 'body', 'image');

}