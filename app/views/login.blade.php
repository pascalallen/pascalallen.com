@extends('layouts.master')

@section('top-script')
	{{-- CUSTOM CSS --}}
	<link rel="stylesheet" type="text/css" href="/css/login.css">
	{{-- CUSTOM FONT --}}
	<link href='https://fonts.googleapis.com/css?family=Ubuntu' rel='stylesheet' type='text/css'>
	{{-- FONT AWESOME --}}
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/font-awesome/4.5.0/css/font-awesome.min.css">
@stop

@section('content')
	<div class="container">
		<div class="header">Bloggy Blog</div>
		<div class="subheader">Log In</div>
		<hr>

		{{ Form::open(array('action' => 'HomeController@postLogin')) }}

			<div class="form-group {{ ($errors->has('email')) ? 'has-error' : '' }}">
				{{ $errors->first('email', '<div class="alert alert-danger">:message</div>') }}
				{{ Form::label('email', 'Email') }}
				{{ Form::text('email', Input::old('email'), array('class' => 'form-control')) }}
			</div>

			<div class="form-group {{ ($errors->has('password')) ? 'has-error' : '' }}">
				{{ $errors->first('password', '<div class="alert alert-danger">:message</div>') }}
				{{ Form::label('password', 'Password') }}
				{{ Form::password('password', array('class' => 'form-control')) }}
			<button type="submit" class="btn btn-default">Submit</button>
			</div>


		{{ Form::close() }}

	</div>
@stop