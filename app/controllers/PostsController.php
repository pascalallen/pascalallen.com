<?php

class PostsController extends \BaseController {

	public function __construct()
	{
		parent::__construct();

		$this->beforeFilter('auth', array('except' => array('index', 'show')));
	}
	/**
	 * Display a listing of the resource.
	 *
	 * @return Response
	 */
	public function index()
	{
		$search = Input::get('search');
		
		if ($search) {
			$query = Post::with('user')->where('title', 'LIKE', '%' . $search . '%')->orWhere('body', 'LIKE', '%' . $search . '%')->orWhere('location', 'LIKE', '%' . $search . '%');
		} else {
			$query = Post::with('user');
		}

		$posts = $query->orderBy('created_at', 'desc')->paginate(30);

		return View::make('posts.index')->with(['posts' => $posts, 'search' => $search]);
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
		if(Auth::id()!= $post->user_id){
			return Redirect::action('posts.index');
		}
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
		Session::flash('successMessage', 'Your post has been deleted.');
		return Redirect::action('posts.index');
	}

	protected function validateAndSave($post)
	{
		$validator = Validator::make(Input::all(), Post::$rules);

		if ($validator->fails()) {
			dd($validator->messages());
	        // validation failed, redirect to the tutorial create page with validation errors and old inputs
	        return Redirect::back()->withInput()->withErrors($validator);
	    } else {
	    	
			$post->title = Input::get('title');
			$post->body = Input::get('body');
			$post->location = Input::get('location');
	    	$image = Input::file('image');

			if ($image) {
				$filename = $image->getClientOriginalName();
				$post->image = '/uploaded/' . $filename;
				$image->move('uploaded/', $filename);
			}
			
			$post->user_id = Auth::id();

			$result = $post->save();

			if($result) {
				Session::flash('successMessage', 'Your post has been saved.');
				return Redirect::action('PostsController@show', $post->id);
			} else {
				return Redirect::back()->withInput();
			}
		}
	}
}
