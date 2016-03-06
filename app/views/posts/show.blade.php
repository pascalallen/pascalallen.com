@extends('layouts.master')

@section('top-script')

	<style type="text/css">
		.row {
			text-align: center;
		}

		.post-image {
			height: 300px;
		}

		a:link {
		    text-decoration: none;
		    color: #777;
		}

		a:visited {
		    text-decoration: none;
		    color: #777;
		}

		a:hover {
		    text-decoration: underline;
		    color: #777;
		}

		a:active {
		    text-decoration: underline;
		    color: #777;
		}
	</style>

@stop

@section('content')

	<div class="row">
		<div class="col-md-8 col-md-offset-2">
			@if(Auth::user() == $post->user)
				<a href="{{{ action('PostsController@edit', $post->id) }}}">{{ Form::submit('Edit', ['class' => 'btn btn-warning']) }}</a>
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
				<footer>Created by <a href="{{{ action('UsersController@show', $post->user->id)}}}">{{{ $post->user->username }}}</a>, {{{$post->created_at->diffForHumans() }}}</footer>
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