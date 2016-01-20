@extends('layouts.master')

@section('top-script')

	{{-- CUSTOM CSS --}}
	<link rel="stylesheet" type="text/css" href="/css/posts-edit.css">
	
@stop

@section('content')
	<div class="container">
		<div class="header">Bloggy Blog</div>
		<div class="subheader">Blog Stuffz</div>
		<hr>

		{{ Form::model($post, array('action' => array('PostController@update', $post->id), 'method' => 'PUT', 'files' => true)) }}

			<div class="form-group {{ ($errors->has('title')) ? 'has-error' : '' }}">
				{{ $errors->first('title', '<div class="alert alert-danger">:message</div>') }}
				{{ Form::label('title', 'Title') }}
				<input type="text" class="form-control" id="title" name="title" value="{{{ $post->title }}}">
			</div>

			<div class="form-group {{ ($errors->has('body')) ? 'has-error' : '' }}">
				{{ $errors->first('body', '<div class="alert alert-danger">:message</div>') }}
				{{ Form::label('body', 'Body') }}
				<input type="text" class="form-control" id="body" name="body" value="{{{ $post->body }}}"></input>
			</div>

			<div class="form-group {{ ($errors->has('image')) ? 'has-error' : '' }}">
				{{ $errors->first('image', '<div class="alert alert-danger">:message</div>') }}
				{{ Form::label('image', 'Image') }}
				{{{ $post->image }}}
				{{ Form::file('image') }}
			</div>

			<button type="submit" class="btn btn-default">Submit</button>
			
		{{ Form::close() }}
	</div>
@stop