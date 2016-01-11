@extends('layouts.master')

@section('content')

<div class="container">
	<h2>Blog Posts</h2>
	<p>Check it out!</p>            
	<table class="table table-condensed">
		<thead>
			<tr>
				<th>Title</th>
				<th>Body</th>
				<th>Id</th>
			</tr>
		</thead>
		<tbody>
				@foreach ($posts as $post)
					<tr>
						<td><a href="{{{ action('PostController@show', $post->id) }}}">{{{ $post->title }}}</a></td>
						<td>{{{ $post->body }}}</td>
						<td>{{{ $post->id }}}</td>
					</tr>
				@endforeach
		</tbody>
	</table>
</div>

@stop