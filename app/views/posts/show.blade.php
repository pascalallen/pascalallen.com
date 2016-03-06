@extends('layouts.master')

@section('top-script')

	<style type="text/css">
		.row {
			text-align: center;
		}

		.post-image {
			height: 300px;
		}
	</style>

@stop

@section('content')

	<div class="row">
		<div class="col-md-8 col-md-offset-2">
			@if(Auth::user() == $post->user)
				{{ Form::submit('Edit', ['class' => 'btn btn-warning']) }}
				{{ Form::open(array('action' => array('PostsController@destroy', $post->id, 'files' => true), 'method' => 'DELETE')) }}
					{{ Form::submit('Delete', ['class' => 'btn btn-danger']) }}
				{{ Form::close() }}
			@endif
			<h3 class="title">{{{ ($post->title) }}}</h3>
			<h5 class="location">{{{ ($post->location) }}}</h3>
			<p class="body">{{{ ($post->body) }}}</p>
			@if (isset($post->image))
				<img src="{{{ $post->image }}}" class="post-image">
			@endif
			<blockquote>
				<footer>Created by {{{ $post->user->username }}}, {{{$post->created_at->diffForHumans() }}}</footer>
			</blockquote>
		</div>
	</div>

@stop

@section('bottom-script')

	<script type="text/javascript">
		$('#edit').click(function() {
			window.location="{{{ action('PostsController@edit', $post->id) }}}";
		});
	</script>

@stop