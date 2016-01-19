@extends('layouts.master')

@section('top-script')

	{{-- CUSTOM CSS --}}
	<link rel="stylesheet" type="text/css" href="/css/posts-index.css">

@stop

@section('content')
<div class="container">
	<div class="row">
		<div class="col-md-8 col-md-offset-2">
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
	</div>
</div>

@stop