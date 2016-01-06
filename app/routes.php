<?php

/*
|--------------------------------------------------------------------------
| Application Routes
|--------------------------------------------------------------------------
|
| Here is where you can register all of the routes for an application.
| It's a breeze. Simply tell Laravel the URIs it should respond to
| and give it the Closure to execute when that URI is requested.
|
*/

Route::get('/', function()
{
	return View::make('hello');
});

Route::get('/resume', function()
{
    return "This is my resume";
});

Route::get('/portfolio', function()
{
    return "This is my portfolio";
});

Route::get('user/{name?}', function($name = 'Pascal')
{
    return $name;
});

Route::get('/sayhello/{name?}', function($name = 'Pascal')
{
    if ($name == "Chris") {
        return Redirect::to('/');
    } else {
    	$data = array('name' => $name);
        return View::make('my-first-view')->with($data);
    }
});

Route::get('/rolldice/{guess}', function($guess)
{	
	$randNum = rand(1, 6);
	$data = array(
		'guess' => $guess, 
		'randNum' => $randNum
	);
    return View::make('roll-dice')->with($data);
});






