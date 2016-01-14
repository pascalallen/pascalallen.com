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

Route::get('/',			 'HomeController@showWelcome');

Route::get('/resume', 	 'HomeController@showResume');

Route::get('/portfolio', 'HomeController@showPortfolio');

Route::get('login', 	 'HomeController@getLogin');

Route::post('login', 	 'HomeController@postLogin');

Route::get('logout', 	 'HomeController@getLogout');

// Route::get('/posts/my-posts/{username}', 'PostsController@showAuthorPosts'); // to create and route a new method

// RESOURCE GOES LAST
Route::resource('/posts', 'PostController');


