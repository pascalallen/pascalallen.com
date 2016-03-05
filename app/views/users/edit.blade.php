@extends('layouts.master')

@section('top-script')

	<style type="text/css">
		.row {
			text-align: center;
		}
	</style>

@stop

@section('content')

	<div class="row">
		<div class="col-md-4 col-md-offset-4">
			{{ Form::model($user, array('action' => array('UsersController@update', $user->id), 'method' => 'PUT', 'files' => true)) }}

				<div class="form-group {{ ($errors->has('username')) ? 'has-error' : '' }}">
					{{ $errors->first('username', '<div class="alert alert-danger">:message</div>') }}
					{{ Form::label('username', 'Username') }}
					{{ Form::text('username', null, ['class' => 'form-control', 'placeholder' => 'Username', 'value' => '{{{ $user->username }}}']) }}
				</div>

				<div class="form-group {{ ($errors->has('password')) ? 'has-error' : '' }}">
					{{ $errors->first('password', '<div class="alert alert-danger">:message</div>') }}
					{{ Form::label('password', 'Password') }}
					{{ Form::password('password', ['class' => 'form-control', 'placeholder' => 'Password']) }}
				</div>

				<div class="form-group {{ ($errors->has('passwordmatch')) ? 'has-error' : '' }}">
					{{ $errors->first('passwordmatch', '<div class="alert alert-danger">:message</div>') }}
					{{ Form::label('passwordmatch', 'Password Match') }}
					{{ Form::password('passwordmatch', ['class' => 'form-control', 'placeholder' => 'Password Again']) }}
				</div>

				<div class="form-group {{ ($errors->has('email')) ? 'has-error' : '' }}">
					{{ $errors->first('email', '<div class="alert alert-danger">:message</div>') }}
					{{ Form::label('email', 'Email') }}
					{{ Form::email('email', null, ['class' => 'form-control', 'placeholder' => 'Email', 'value' => '{{{ $user->email }}}']) }}
				</div>

				<div class="form-group {{ ($errors->has('image')) ? 'has-error' : '' }}">
					{{ $errors->first('image', '<div class="alert alert-danger">:message</div>') }}
					{{ Form::label('image', 'Image') }}
					<p><u>Current Image:</u> {{{ $user->image }}}</p>
					{{ Form::file('image') }}
				</div>

				<div class="{{ ($errors->has('location')) ? 'has-error' : '' }} form-group">
					{{ $errors->first('location', '<div class="alert alert-danger">:message</div>') }}
					{{ Form::label('location', 'Location') }}
					{{ Form::text('location', null, ['class' => 'form-control', 'placeholder' => 'Location', 'value' => '{{{ $user->location }}}']) }}
				</div>

				{{ Form::submit('submit', ['class' => 'btn btn-default']) }}	
			{{ Form::close() }}
		</div>
	</div>

@stop

@section('bottom-script')

@stop