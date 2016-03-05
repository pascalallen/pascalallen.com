<?php

use Illuminate\Auth\UserTrait;
use Illuminate\Auth\UserInterface;
use Illuminate\Auth\Reminders\RemindableTrait;
use Illuminate\Auth\Reminders\RemindableInterface;

class User extends Eloquent implements UserInterface, RemindableInterface {

	use UserTrait, RemindableTrait;

	/**
	 * The database table used by the model.
	 *
	 * @var string
	 */
	protected $table = 'users';

	/**
	 * The attributes excluded from the model's JSON form.
	 *
	 * @var array
	 */

	public static $rules = array(
		    'username'  => 'required|max:100',
		    'email'     => 'required|max:100',
		    //look at https://github.com/esensi/model#validating-model-trait
		    'password'	=> 'required|max:100|min:6',
		    'image' 	=> 'image',
		    'location'	=> 'max:100|min:6'
		);

	protected $hidden = array('password', 'remember_token');

	public function setPasswordAttribute($value)
    {
    	$this->attributes['password'] = Hash::make($value);
    }

    public function post()
	{
	    return $this->hasMany('Post');
	}

}
