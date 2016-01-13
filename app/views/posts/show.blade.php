@extends('layouts.master')

@section('top-script')
	{{-- CUSTOM CSS --}}
	<link rel="stylesheet" type="text/css" href="/css/posts-show.css">
	{{-- CUSTOM FONT --}}
	<link href='https://fonts.googleapis.com/css?family=Ubuntu' rel='stylesheet' type='text/css'>

	{{-- FONT AWESOME --}}
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/font-awesome/4.5.0/css/font-awesome.min.css">
@stop

@section('content')
	<div class="container">
		<div class="header">Bloggy Blog</div>
		<div class="subheader">Blog Stuffz</div>
		<hr>

		{{ Form::open(array('action' => array('PostController@destroy', $post->id), 'method' => 'DELETE')) }}
		
			<a href="{{{ action('PostController@edit', $post->id) }}}" class="btn btn-info">Edit</a>

			<button class="btn btn-danger">Delete</button>

		{{ Form::close() }}

		<h1>{{{ $post->title }}}</h1>
		<p>{{{ $post->body }}}</p>
	</div>
@stop