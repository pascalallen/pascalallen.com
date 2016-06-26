<?php

class TattooImagesController extends \BaseController {

	/**
	 * Display a listing of tattooimages
	 *
	 * @return Response
	 */
	public function index()
	{
		$tattooimages = TattooImage::orderBy('created_at', 'desc')->paginate(10);

		return View::make('tattoo-artist-portfolio')->with(['tattooimages' => $tattooimages]);
	}


	/**
	 * Show the form for creating a new tattooimage
	 *
	 * @return Response
	 */
	public function create()
	{
		return View::make('tattooimages.create');
	}

	/**
	 * Store a newly created tattooimage in storage.
	 *
	 * @return Response
	 */
	public function store()
	{
		$validator = Validator::make($data = Input::all(), Tattooimage::$rules);

		if ($validator->fails())
		{
			return Redirect::back()->withErrors($validator)->withInput();
		}

		Tattooimage::create($data);

		return Redirect::route('tattooimages.index');
	}

	/**
	 * Display the specified tattooimage.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function show($id)
	{
		$tattooimage = Tattooimage::findOrFail($id);

		return View::make('tattooimages.show', compact('tattooimage'));
	}

	/**
	 * Show the form for editing the specified tattooimage.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function edit($id)
	{
		$tattooimage = Tattooimage::find($id);

		return View::make('tattooimages.edit', compact('tattooimage'));
	}

	/**
	 * Update the specified tattooimage in storage.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function update($id)
	{
		$tattooimage = Tattooimage::findOrFail($id);

		$validator = Validator::make($data = Input::all(), Tattooimage::$rules);

		if ($validator->fails())
		{
			return Redirect::back()->withErrors($validator)->withInput();
		}

		$tattooimage->update($data);

		return Redirect::route('tattooimages.index');
	}

	/**
	 * Remove the specified tattooimage from storage.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function destroy($id)
	{
		Tattooimage::destroy($id);

		return Redirect::route('tattooimages.index');
	}

}
