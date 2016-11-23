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
			{!! Form::open(array('action' => 'UsersController@store', 'method' => 'post', 'files' => true)) !!}
				<div class="{!! ($errors->has('name')) ? 'has-error' : '' !!} form-group">
					{!! $errors->first('name', '<div class="alert alert-danger">:message</div>') !!}
					{!! Form::label('name', 'Name') !!}
					{!! Form::text('name', null, ['class' => 'form-control', 'placeholder' => 'Name']) !!}
				</div>
				<div class="{!! ($errors->has('password')) ? 'has-error' : '' !!} form-group">
					{!! $errors->first('password', '<div class="alert alert-danger">:message</div>') !!}
					{!! Form::label('password', 'Password') !!}
					{!! Form::password('password', ['class' => 'form-control', 'placeholder' => 'Password']) !!}
				</div>
				<div class="{!! ($errors->has('email')) ? 'has-error' : '' !!} form-group">
					{!! $errors->first('email', '<div class="alert alert-danger">:message</div>') !!}
					{!! Form::label('email', 'Email') !!}
					{!! Form::email('email', null, ['class' => 'form-control', 'placeholder' => 'Email']) !!}
				</div>
				{!! Form::submit('submit', ['class' => 'btn btn-default']) !!}
			{!! Form::close() !!}
			<a href="{!! action('HomeController@getLogin') !!}">Already a member? Login!</a>
		</div>
	</div>

@stop

@section('bottom-script')

@stop