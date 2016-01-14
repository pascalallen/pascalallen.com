@extends('layouts.master')

@section('top-script')
	{{-- CUSTOM CSS --}}
	<link rel="stylesheet" type="text/css" href="/css/posts-index.css">
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
	<table class="table table-nonfluid center-table">
		<thead>
			<tr>
				<th>Title</th>
				<th>Body</th>
				<th>Image</th>
				<th>Created</th>
			</tr>
		</thead>
		<tbody>
				@foreach ($posts as $post)
					<tr>
						<td><a href="{{{ action('PostController@show', $post->id) }}}" class="posts-title">{{{ $post->title }}}</a></td>
						<td>{{{ $post->body }}}</td>
						<td><img src="{{{ $post->image }}}" class="post-image"></td>
						<td>{{{ $post->created_at->diffForHumans() }}}</td>
					</tr>
				@endforeach
		</tbody>
	</table>
	{{ $posts->links() }}
</div>

@stop