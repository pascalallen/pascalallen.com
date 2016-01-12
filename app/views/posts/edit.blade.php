@extends('layouts.master')

@section('top-script')
	{{-- CUSTOM CSS --}}
	<link rel="stylesheet" type="text/css" href="/css/posts-edit.css">
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

		{{ Form::model($post, array('action' => array('PostController@update', $post->id), 'method' => 'PUT')) }}

			<div class="form-group {{ ($errors->has('title')) ? 'has-error' : '' }}">
				{{ $errors->first('title', '<div class="alert alert-danger">:message</div>') }}
				<label for="title">Title: </label>
				<input type="text" class="form-control" id="title" name="title" value="{{{ $post->title }}}">
			</div>

			<div class="form-group {{ ($errors->has('body')) ? 'has-error' : '' }}">
				{{ $errors->first('body', '<div class="alert alert-danger">:message</div>') }}
				<label for="body">Body: </label>
				<input type="text" class="form-control" id="body" name="body" value="{{{ $post->body }}}"></input>
			</div>

			<div class="form-group {{ ($errors->has('image')) ? 'has-error' : '' }}">
				{{ $errors->first('image', '<div class="alert alert-danger">:message</div>') }}
				<label for="image">Image: </label>
				<input type="file" id="image" name="image" value="{{{ $post->image }}}">
			</div>

			<button type="submit" class="btn btn-default">Submit</button>
		{{ Form::close() }}
	</div>
@stop