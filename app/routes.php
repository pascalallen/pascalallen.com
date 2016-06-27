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

Route::get('login', 'HomeController@getLogin');

Route::post('login', 'HomeController@postLogin');

Route::get('logout', 'HomeController@getLogout');

Route::get('/search', 'HomeController@search');

Route::get('search-show/{id}', 'HomeController@searchShow');

Route::get('random-student', 'HomeController@randomStudent');

Route::get('tattoo-artist-project', 'HomeController@tattooArtistProject');

Route::get('tattoo-artist-portfolio', 'TattooImagesController@index');

Route::get('coleman', 'HomeController@coleman');

Route::get('/resume', 	 'HomeController@showResume');

Route::get('/portfolio', 'HomeController@showPortfolio');

Route::get('/my_calculator', 'HomeController@showCalculator');

Route::get('/texter', 'HomeController@showTexter');

Route::resource('national_parks', 'NationalParksController');

Route::resource('tattoo_images', 'TattooImagesController');

Route::resource('posts', 'PostsController');

Route::resource('users', 'UsersController');




