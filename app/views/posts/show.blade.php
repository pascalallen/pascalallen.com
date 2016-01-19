@extends('layouts.master')

@section('top-script')

	{{-- CUSTOM CSS --}}
	<link rel="stylesheet" type="text/css" href="/css/posts-show.css">

@stop

@section('content')
	<div class="container">
		<div class="header">Bloggy Blog</div>
		<div class="subheader">Blog Stuffz</div>
		<hr>

		{{ Form::open(array('action' => array('PostController@destroy', $post->id, 'files' => true), 'method' => 'DELETE')) }}
		
			<a href="{{{ action('PostController@edit', $post->id) }}}" class="btn btn-info">Edit Post</a>

			<a href="{{{ action('PostController@editImage', $post->id) }}}" class="btn btn-info">Edit Image</a>

			<button class="btn btn-danger">Delete</button>

		{{ Form::close() }}

		<h1>{{{ $post->title }}}</h1>
		<p>{{{ $post->body }}}</p>
		<img src="{{{ $post->image }}}" class="post-image">
	</div>
@stop