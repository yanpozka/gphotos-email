import React, { Component } from 'react';
import { StyleSheet, WebView } from 'react-native';
import { Body, Button, Container, Content, Header, Icon, List, ListItem, Text, Thumbnail } from 'native-base';
import { StackNavigator } from 'react-navigation';

import { PhotosList } from './photoslist'
import config from './config';


class Main extends Component {
	constructor(props) {
		super(props);
		this.state = { isLoading: true };
	}

	render() {
		return (<PhotosList />)

		/*const { navigate } = this.props.navigation;
		return (
		  <Container>
			<Header />
			<Button onPress={() => {
			  navigate('oauth');
			}}>
			  <Icon name='logo-google' />
			  <Text>Login with Google</Text>
			</Button>
		  </Container>
		)*/
	}
}

class OauthWebView extends Component {
	static navigationOptions = {
		title: 'Login with your Google account'
	};

	constructor(props) {
		super(props);
		this.state = {
			isLoading: true,
			url: null,
		};
	}

	componentDidMount() {
		const self = this;

		fetch(config.url + '/loginurl')
			.then(function (response) { return response.json(); })
			.then(function (urlObj) {
				self.setState({ isLoading: false, url: urlObj.url });
			})
			.catch(function (err) {
				console.error('[-] Caught error: ', err);
			})
	}

	render() {
		if (this.state.isLoading) {
			return (
				<Container><Text>Loading ...</Text></Container>
			)
		}

		return (
			<WebView
				source={{ uri: this.state.url }}
				// TODO: open the default brower
				userAgent="Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3163.100 Safari/537.36"
			/>
		)
	}
}

const App = StackNavigator({
	login: { screen: Main },
	oauth: { screen: OauthWebView },
	listPhotos: { screen: PhotosList }
});

export default class Photos4MartaApp extends React.Component {
	render() {
		return <App />;
	}
}

const styles = StyleSheet.create({
	// const styles = {
	itemList: {
		backgroundColor: '#FFFFFF'
	},
	itemListSelected: {
		backgroundColor: '#887C7A'
	},
});