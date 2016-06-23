<?php
	// FIRST RUN MIGRATION AND SEEDER FILE.
	require_once "../parks_logins.php";
	require_once "../db_connect.php";
	require_once "../park_migration.php";
	require_once "../Input.php";

	function checkValues()
	{
		return Input::setAndNotEmpty('name') && Input::setAndNotEmpty('location') && Input::setAndNotEmpty('date_established') && Input::setAndNotEmpty('area_in_acres') && Input::setAndNotEmpty('description');
	}

	function insertPark($dbc)
	{
		$errors = [];

		try{
			$name = Input::has('name') ? Input::getString('name') : null;
		} catch (Exception $e) {
			array_push($errors, $e->getMessage());
		}
		try{
			$location = Input::has('location') ? Input::getString('location') : null;
		} catch (Exception $e) {
			array_push($errors, $e->getMessage());
		}
		try{
			$date_established = Input::has('date_established') ? Input::getDate('date_established') : null;
			$date_established = $date_established->format("Y-m-d");
		} catch (Exception $e) {
			array_push($errors, $e->getMessage());
		}
		try{
			$area_in_acres = Input::has('area_in_acres') ? Input::getNumber('area_in_acres') : null;
		} catch (Exception $e) {
			array_push($errors, $e->getMessage());
		}
		try{
			$description = Input::has('description') ? Input::getString('description') : null;
		} catch (Exception $e) {
			array_push($errors, $e->getMessage());
		}

		if(!empty($errors)){
			return $errors;
		}


		$insert_table = "INSERT INTO national_parks (name, location, date_established, area_in_acres, description) VALUES (:name, :location, :date_established, :area_in_acres, :description)";

	    $stmt = $dbc->prepare($insert_table);
	    $stmt->bindValue(':name', $name, PDO::PARAM_STR);
	    $stmt->bindValue(':location', $location, PDO::PARAM_STR);
	    $stmt->bindValue(':date_established', $date_established, PDO::PARAM_STR);
	    $stmt->bindValue(':area_in_acres', $area_in_acres, PDO::PARAM_STR);
	    $stmt->bindValue(':description', $description, PDO::PARAM_STR);

	    $stmt->execute();

		return $errors;
	}


	function deletePark($dbc)
	{
		if (Input::has('id')) {
			$delete_column = "DELETE FROM national_parks WHERE id = :id";
			$del = $dbc->prepare($delete_column);
			$del->bindValue(':id', Input::get('id'), PDO::PARAM_STR);
			$del->execute();
		}
	}

	function pageController($dbc)
	{
		$errors = null;


		if (!empty($_POST)) {
			if (checkValues()) {
				$errors = insertPark($dbc);			
			} else {
				$message = "Invalid format. Please try again.";
				$javascript = "<script type='text/javascript'>alert('$message');</script>";
				echo $javascript;
			}
		}

		deletePark($dbc);

		// Count
		$countAll = 'SELECT count(*) FROM national_parks';
		$count_stmt = $dbc->query($countAll);
		$count = $count_stmt->fetchColumn();
		$limit = 2;
		$max_page = ceil($count / $limit);

		// Sanitizing
		$page = Input::has('page') ? Input::get('page') : 1; // grabs url value if exists, if not set to 1
		$page = (is_numeric($page)) ? $page : 1; // is value numeric, if not set to 1
		$page = ($page > 0) ? $page : 1; // is value greater than zero, if not set to 1
		$page = ($page <= $max_page) ? $page : $max_page; // is value less than or equal maximum amount of pages, if not set to max page

		// Offset
		$offset = $page * $limit - $limit;
		$selectAll = 'SELECT * FROM national_parks LIMIT :limit OFFSET :offset';
		$stmt = $dbc->prepare($selectAll);
		$stmt->bindValue(':limit', $limit, PDO::PARAM_INT);
		$stmt->bindValue(':offset', $offset, PDO::PARAM_INT);
		$stmt->execute();
		$parks = $stmt->fetchAll(PDO::FETCH_ASSOC);

		return array(
			'page' => $page,
			'parks' => $parks,
			'errors' => $errors,
			'max_page' => $max_page
			);
	}
	extract(pageController($dbc));
?>
<!doctype html>
<html>
<head>
	<meta charset="UTF-8">
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
	<title>National Parks</title>
	<!-- BOOTSTRAP -->
	<link href="css/bootstrap.min.css" rel="stylesheet">
	<!-- CUSTOM CSS -->
	<link rel="stylesheet" type="text/css" href="css/national_parks.css">
	<!-- TITLE IMG -->
	<link rel="shortcut icon" href="img/codeup-arrow.png">
	<!-- CUSTOM FONT -->
	<link href='https://fonts.googleapis.com/css?family=Rock+Salt' rel='stylesheet' type='text/css'>
</head>
<body>
	<div class="site-wrapper">
		<div class="container">
			<!-- <h1><?= $page; ?></h1> -->
			<h2>National Parks</h2>
			<!-- <h3>Database Driven Web Application</h3> -->
			<div class="row">
				<div class="col-xs-8 col-xs-offset-2">
					<table class="table">
						<thead>
							<tr>
							<th>Name</th>
							<th>Location</th>
							<th>Established</th>
							<th>Acres</th>
							<th>Description</th>
							<th></th>
							</tr>
					    </thead>
						<?php foreach ($parks as $park) : ?>
							<tbody>
								<tr class="table-bordered">
								    <td><?= Input::escape($park['name']) ?> </td>
								    <td><?= Input::escape($park['location']) ?> </td>
								    <td><?= Input::escape($park['date_established']) ?> </td>
								    <td><?= Input::escape($park['area_in_acres']) ?> </td>
								    <td><?= Input::escape($park['description']) ?> </td>
								    <td>
							    		<form role="form" method="POST">
											<button type="submit" class="btn btn-info btn-xs" value="<?= $park['id'] ?>" name="id">Delete</button>
										</form>
									</td>
								</tr>
							</tbody>
						<?php endforeach; ?>
					</table>
					<?php if ($page != 1) : ?>
						<a href="?page=<?= ($page - 1) ?>">Previous</a>
					<?php endif; ?>
					<?php if ($page != $max_page) : ?>
						<a href="?page=<?= ($page + 1) ?>">Next</a>
					<?php endif; ?>
				</div>
			</div>

			<div class="row">
				<h3 class="col-xs-4 col-xs-offset-4">Submit a park:</h3>
				<div class="col-xs-4 col-xs-offset-4">
					<?php if(!empty($errors)) : ?>
						<?php foreach($errors as $error) : ?>
							<h5> <?= $error ?> </h5>
						<?php endforeach ?>
					<?php endif ?>
					<form role="form" method="POST">
						<div class="form-group">
							<label for="name">Name:</label>
							<input type="text" class="form-control" name="name" id="name" placeholder="Enter name">
						</div>
						<div class="form-group">
							<label for="location">Location:</label>
							<input type="text" class="form-control" name="location" id="location" placeholder="Enter location">
						</div>
						<div class="form-group">
							<label for="date">Date Established:</label>
							<input type="text" class="form-control" name="date_established" id="date" placeholder="Enter date established">
						</div>
						<div class="form-group">
							<label for="acres">Acres:</label>
							<input type="text" class="form-control" name="area_in_acres" id="acres" placeholder="Enter acres">
						</div>
						<div class="form-group">
							<label for="description">Description:</label>
							<input type="text" class="form-control" name="description" id="description" placeholder="Enter description">
						</div>
						<button type="submit" class="btn btn-default">Submit</button>
					</form>
				</div>
			</div>
		</div>
	</div>
		<!-- JQUERY -->
		<script src="js/jquery-2.1.4.min.js"></script>
	<!-- CUSTOM JAVASCRIPT -->
	<script src="js/national_parks.js"></script>
</body>
</body>
</html>





