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
		<div class="col-md-8 col-md-offset-2">
			@if(Auth::user() != $user)
				{!! $user->name !!}
			@else
				<h1>Let's do this, {!! $user->name !!}!</h1>
				<br>
			@endif
			<br>
			<h1>{!! $user->name !!}'s Content</h1>
			blah
			blah
			blah
			blah
		</div>
	</div>
@stop

@section('bottom-script')

@stop