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

Route::get('/', 'HomeController@showWelcome');

Route::get('user/{name?}', 'HomeController@showName');

Route::get('/sayhello/{name?}', 'HomeController@sayHello');

Route::get('/rolldice/{guess?}', 'HomeController@rollDice');

Route::get('/resume', 'HomeController@showResume');

Route::get('/portfolio', 'HomeController@showPortfolio');

Route::resource('/posts', 'PostsController');

Route::get('/posts/my-posts/{username}', 'PostsController@showAuthorPosts');