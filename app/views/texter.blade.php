<!DOCTYPE html>
<html>
	<head>
		<title>Kandy SMS Texter</title>
		<meta content="width=device-width, initial-scale=1.0" name="viewport">

		<!-- BOOTSTRAP -->
		<link href="css/bootstrap.min.css" rel="stylesheet">

		<!--Load Pace AJAX Progress Bar (optional)-->
		<script src="https://cdnjs.cloudflare.com/ajax/libs/pace/1.0.2/pace.min.js"></script>
		<link rel="stylesheet" media="screen" href="https://cdnjs.cloudflare.com/ajax/libs/pace/1.0.2/themes/pink/pace-theme-minimal.css">

		<!-- Custom Font -->
		<link href='https://fonts.googleapis.com/css?family=Poiret+One' rel='stylesheet' type='text/css'>

		<!-- Custom CSS -->
		<link href="css/smssend.css" rel="stylesheet">

		<!-- TITLE IMG -->
		<link rel="shortcut icon" href="img/kandy_icon.jpg">

	</head>
	<body>
		<div class="container">
			<div class="row">
				<div class="col-xs-8 col-xs-offset-2" id="activity-container">
					<div class="clearfix">
						<p class="h4 pull-left">
							<strong><span class="username"></span></strong>
						</p>
					</div>
					<div id="sms-container">
						<h1 class="header">Kandy SMS Texter</h1>
						{{-- <div id="chat-messages" value=""></div> --}}
						<div id="sms-messages"></div>
						<div id="sms-input">
							<div class="form-group">
								<label for="sms_to">Receiver</label>
								<input type="text" name="sms_to" id="sms_to" class="form-control" placeholder="Friends cell #">
							</div>
							<div class="form-group">
								<label for="sms_message">Message</label>
								<textarea type="text" name="sms_message" id="sms_message" class="form-control" placeholder="Your message"></textarea>
							</div>
							<button name="button" type="submit" class="btn btn-success" id="send-btn">Send</button>
						</div>
					</div>
				</div>
			</div>
		</div>
		<!--Load Kandy JS files -->
		<script src="https://code.jquery.com/jquery-1.7.2.min.js"></script>
		<script src="https://kandy-portal.s3.amazonaws.com/public/javascript/kandy/2.3.0/kandy.js"></script>
		<!-- Custom JS -->
		<script src="js/smssend.js"></script>
	</body>
</html>
