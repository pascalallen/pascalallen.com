App/commands

Laravel allows you to write custom commands. Any custom commands that you write should be saved in this directory.

App/config

Laravel configuration files go here. Things like your database connection info and any other configuration items will go in here. We will talk more about the specifics when we configure our blog application.

App/controllers

Remember the 'C' in MVC? The controllers for your application go in this directory.

App/database

Database migrations and seed files go in this directory. We will discuss these more later, but migrations and seeds allow you to create database tables and insert records into the tables.

App/lang

Laravel supports localization. Language files can be placed in this directory. We will not be using this feature in our blog application.

App/models

Remember the 'M' in MVC? The models for your application go in this directory.

App/start

Files that get loaded and process on startup are found in this directory. The primary reasons to make changes in this directory are to add additional directories to the autoloader path or to customize how errors are handled by your application.

App/storage

This is the directory where Laravel stores log files, temp files, sessions, etc. You will primarily access this directory to view the Laravel log file.

App/tests

Your application tests go in this directory. Laravel provides some example test cases for you.

App/views

Remember the 'V' in MVC? The views for your application go in this directory.

App/filters.php

Laravel provides something called filters that can be applied to a route to change the way it behaves. This will be discussed in more detail later.

App/routes.php

Remember how we discussed that when you access a URL through Laravel that you won't see something like blog.dev/post.php? Instead, Laravel will route requests based on paths that are configured in this file. You will be accessing and changing this file often, so do not forget where to find it.


I AM USING LARAVEL VERSION 4.2 https://laravel.com/docs/4.2
READ THIS BOOK http://daylerees.com/codebright/
PACKAGIST.ORG IS A PHP FRAMEWORK REPOSITORY SITE — WHERE LARAVEL FRAMEWORK LIVES
ALTER HELLO.PHP FILE UNDER BLOG.DEV

## >COMPOSER REQUIRE FOLDER/FILE
## artisan is a command line interface to your laravel application
## php artisan env (environment)
## php artisan routes (produces routes(get/post requests))
## REPL for laravel php artisan tinker
## ctrl+D to exit tinker
## Everything we do will be mostly in models, controllers, views
## css js img go in public

## BLADE
---------------------------------
{{{ $SOMETHING }}} // to protect code using blade

@if($name == ’something’) // if statement in blade
	<h1>soasdfasd</h1>
@endif

@foreach ($cohorts as $className) //foreach in blade
	<h3>{{{ $className }}}</h3>
@endforeach

@yield('content') // will go in layout

@extends('layouts.master') // will go in view

@section('content')
	<h1>Guess: {{{ $guess }}} Random Number: {{{ $randNum }}}!</h1>
@stop
	
<!-- CREATE A DATABASE -->
php artisan migrate:install <!-- install the migrations table -->
php artisan migrate:make create_posts_table <!-- create a new migration file -->
class CreatePostsTable extends Migration
{
    public function up()
    {
        Schema::create('posts', function($table)
        {
            $table->increments('id');
            $table->string('title', 100);
            $table->text('body');
            $table->timestamps();
        });
    }

    public function down()
    {
        Schema::drop('posts');
    }
}
php artisan migrate <!-- write and create table -->
php artisan migrate:rollback <!-- rollback a single migration -->
php artisan migrate:reset <!-- rollback and then re-run all migrations -->

## NOTES ON HOW TO PULL PROJECT FROM GITHUB
-- from sites dir
-- git clone //link here //name site
-- ##from vagrant-lamp
-- ansible-playbook ansible/playbooks/vagrant/sites/php.yml
-- composer install from vagrant ssh  site.dev
-- fill out env-template with db credentials

Verb    Path    Controller Method/Action    Description
GET    /posts               index      Show a list of all posts
GET    /posts/create        create     Show a form for creating a post
POST   /posts               store      Store the new post
GET    /posts/{post}        show       Show a specific post
GET    /posts/{post}/edit   edit       Show a form for editing a specific post
PUT    /posts/{post}        update     Update a specific post
DELETE /posts/{post}        destroy    Delete a specific post


## SHOWING OLD INPUT
<input type="text" name="name" value="{{{ Input::old('name') }}}">


## DEFAULT VALUE
$name = Input::get('name', 'Bob');

// would be equivalent to:

$name = 'Bob';
if (isset($_REQUEST['name'])) {
    $name = $_REQUEST['name'];
}

// or:

$name = isset($_REQUEST['name']) ? $_REQUEST['name'] : 'Bob';

## REDIRECT BACK WITH OLD INPUT
return Redirect::back()->withInput();

// or:

return Redirect::action('PostsController@create')->withInput();


#####FORMS
POST
{{ Form::open(array('action' => 'PostsController@store')) }}
{{ Form::close() }}

PUT/DELETE
{{ Form::open(array('action' => array('PostsController@update', $post->id), 'method' => 'PUT')) }}
{{ Form::close() }}

{{ Form::open($post, array('action' => array('PostsController@update', $post->id), 'method' => 'PUT')) }}
{{ Form::close() }}

####FORM INPUTS
--LABEL
{{ Form::label('title', 'Title') }}
--INPUT
{{ Form::text('title', null, ['class' => 'form-control', 'placeholder' => 'Enter your title']) }}




