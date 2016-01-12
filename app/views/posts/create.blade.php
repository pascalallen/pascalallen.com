@extends('layouts.master')

@section('top-script')
	{{-- CUSTOM CSS --}}
	<link rel="stylesheet" type="text/css" href="/css/posts-create.css">
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

		{{ Form::open(array('action' => 'PostController@store')) }}

			<div class="form-group {{ ($errors->has('title')) ? 'has-error' : '' }}">
				{{ $errors->first('title', '<div class="alert alert-danger">:message</div>') }}
				{{ Form::label('title', 'Title') }}
				{{ Form::text('title', null, ['class' => 'form-control', 'placeholder' => 'Enter your title']) }}
			</div>

			<div class="form-group {{ ($errors->has('body')) ? 'has-error' : '' }}">
				{{ $errors->first('body', '<div class="alert alert-danger">:message</div>') }}
				{{ Form::label('body', 'Body') }}
				{{ Form::textarea('body', null, ['class' => 'form-control', 'placeholder' => 'Enter your body']) }}
			</div>

			<div class="form-group {{ ($errors->has('image')) ? 'has-error' : '' }}">
				{{ $errors->first('image', '<div class="alert alert-danger">:message</div>') }}
				{{ Form::label('image', 'Image') }}
				{{ Form::file('image') }}
			</div>

			<button type="submit" class="btn btn-default">Submit</button>
		{{ Form::close() }}
	</div>
@stop