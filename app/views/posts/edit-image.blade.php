@extends('layouts.master')

@section('top-script')

	{{-- CUSTOM CSS --}}
	<link rel="stylesheet" type="text/css" href="/css/posts-edit-image.css">
	
@stop

@section('content')
	<div class="container">
		<div class="header">Bloggy Blog</div>
		<div class="subheader">Blog Stuffz</div>
		<hr>

		{{ Form::model($post, array('action' => array('PostController@update', $post->id), 'method' => 'PUT')) }}

			<div class="form-group {{ ($errors->has('image')) ? 'has-error' : '' }}">
				{{ $errors->first('image', '<div class="alert alert-danger">:message</div>') }}
				{{ Form::label('image', 'Image') }}
				<input type="file" id="image" name="image">
			</div>

			<button type="submit" class="btn btn-default">Submit</button>
			
		{{ Form::close() }}