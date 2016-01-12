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

		{{ Form::open(array('action' => array('PostsController@update', $post->id), 'method' => 'PUT')) }}

			{{ $errors->first('title', '<div class="alert alert-danger">:message</div>') }}

			<div class="form-group {{ ($errors->has('title')) ? 'has-error' : '' }}">
				<label for="title">Title: </label>
				<input type="text" class="form-control" id="title" name="title" value="{{{ Input::old('title') }}}">
			</div>

			{{ $errors->first('body', '<div class="alert alert-danger">:message</div>') }}

			<div class="form-group {{ ($errors->has('body')) ? 'has-error' : '' }}">
				<label for="body">Body: </label>
				<textarea type="text" class="form-control" id="body" name="body" value="{{{ Input::old('body') }}}"></textarea>
			</div>

			{{ $errors->first('user_id', '<div class="alert alert-danger">:message</div>') }}

			<div class="form-group {{ ($errors->has('user_id')) ? 'has-error' : '' }}">
				<label for="user_id">User Id: </label>
				<input type="text" class="form-control" id="user_id" name="user_id" value="1">
			</div>

			{{ $errors->first('image', '<div class="alert alert-danger">:message</div>') }}
			
			<div class="form-group {{ ($errors->has('image')) ? 'has-error' : '' }}">
				<label for="image">Image: </label>
				<input type="file" id="image" name="image" value="{{{ Input::old('image') }}}">
			</div>

			<button type="submit" class="btn btn-default">Submit</button>
		{{ Form::close() }}
	</div>
@stop