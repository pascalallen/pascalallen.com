<?php

class PostController extends \BaseController {

	/**
	 * Display a listing of the resource.
	 *
	 * @return Response
	 */
	public function index()
	{
		$posts = Post::paginate(4);
		// $posts = DB::table('posts')->paginate(4);

		return View::make('posts.index')->with('posts', $posts);
	}


	/**
	 * Show the form for creating a new resource.
	 *
	 * @return Response
	 */
	public function create()
	{
		return View::make('posts.create');
	}


	/**
	 * Store a newly created resource in storage.
	 *
	 * @return Response
	 */
	public function store()
	{
		$post = new Post();
		Session::flash('successMessage', 'Your post has been saved.');
		Log::info(Input::all());
		return $this->validateAndSave($post);
	}


	/**
	 * Display the specified resource.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function show($id)
	{
		$post = Post::find($id);
		if(!$post) {
			App::abort(404);
		}

		return View::make('posts.show')->with('post', $post);
	}


	/**
	 * Show the form for editing the specified resource.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function edit($id)
	{
		$post = Post::find($id);
		return View::make('posts.edit')->with('post', $post);
	}


	/**
	 * Update the specified resource in storage.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function update($id)
	{
		$post = Post::find($id);
		Session::flash('successMessage', 'Your change has been saved.');
		return $this->validateAndSave($post);
	}


	/**
	 * Remove the specified resource from storage.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function destroy($id)
	{
		$post = Post::find($id);
		$post->delete();
		Session::flash('errorMessage', 'Your post has been deleted.');
		return Redirect::action('posts.index');
	}

	protected function validateAndSave($post)
	{
		$validator = Validator::make(Input::all(), Post::$rules);

		if ($validator->fails()) {
	        // validation failed, redirect to the post create page with validation errors and old inputs
	        return Redirect::back()->withInput()->withErrors($validator);
	    } else {
			$post->title = Input::get('title');
			$post->body = Input::get('body');
			$post->image = '/img/header.jpg';
			$post->user_id = 1;

			$result = $post->save();

			if($result) {
				return Redirect::action('PostController@show', $post->id);
			} else {
				return Redirect::back()->withInput();
			}
		}
	}

}
