@extends('layouts.master')

@section('top-script')

	<style type="text/css">
		.row {
			margin: 0;
			text-align: center;
		}

		.create {
			position: absolute;
			margin-top: 80px;
		}

	</style>

@stop

@section('content')

	<div class="row">
		<div class="col-xs-4 col-xs-offset-4 create">
			{{ Form::open(array('action' => 'UsersController@store', 'method' => 'post', 'files' => true)) }}
			
				<div class="{{ ($errors->has('username')) ? 'has-error' : '' }} form-group">
					{{ $errors->first('username', '<div class="alert alert-danger">:message</div>') }}
					{{ Form::label('username', 'Username') }}
					{{ Form::text('username', null, ['class' => 'form-control', 'placeholder' => 'Username']) }}
				</div>

				<div class="{{ ($errors->has('password')) ? 'has-error' : '' }} form-group">
					{{ $errors->first('password', '<div class="alert alert-danger">:message</div>') }}
					{{ Form::label('password', 'Password') }}
					{{ Form::password('password', ['class' => 'form-control', 'placeholder' => 'Password']) }}
				</div>

				<div class="{{ ($errors->has('email')) ? 'has-error' : '' }} form-group">
					{{ $errors->first('email', '<div class="alert alert-danger">:message</div>') }}
					{{ Form::label('email', 'Email') }}
					{{ Form::email('email', null, ['class' => 'form-control', 'placeholder' => 'Email']) }}
				</div>

				<div class="{{ ($errors->has('image')) ? 'has-error' : '' }} form-group">
					{{ $errors->first('image', '<div class="alert alert-danger">:message</div>') }}
					{{ Form::label('image', 'Image') }}
					{{ Form::file('image') }}
				</div>

				<div class="{{ ($errors->has('location')) ? 'has-error' : '' }} form-group">
					{{ $errors->first('location', '<div class="alert alert-danger">:message</div>') }}
					{{ Form::label('location', 'Location') }}
					{{ Form::text('location', null, ['class' => 'form-control', 'placeholder' => 'Location']) }}
				</div>

				{{ Form::submit('submit', ['class' => 'btn btn-default']) }}

			{{ Form::close() }}
		</div>
	</div>

@stop

@section('bottom-script')

@stop