<?php

use Carbon\Carbon;

class Post extends BaseModel
{
    protected $table = 'posts';

    protected $fillable = array('title', 'body');

    public static $rules = array(
	    'title'      => 'required|min:2|max:100',
	    'body'       => 'required|min:2|max:10000',
	    'image'		 => 'image'
	);

    public function user()
	{
	    return $this->belongsTo('User');
	}
}