import React, { Component } from 'react';
import { StyleSheet, WebView } from 'react-native';
import { Body, Button, Container, Content, Header, Icon, List, ListItem, Text, Thumbnail } from 'native-base';
import { StackNavigator } from 'react-navigation';

const config = {
  url: 'http://10.0.0.120:8080',
};

class Main extends Component {
  constructor(props) {
    super(props);
    this.state = { isLoading: true };
  }

  render() {
    // if (logued) { return (<PhotosList />) }
    const { navigate } = this.props.navigation;

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
    )
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
        // TODO: make it work for iOS, this only works for Android :(
        userAgent="Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3163.100 Safari/537.36"
      />
    )
  }
}

class PhotosList extends Component {
  render() {
    var photos = [
      { name: 'Uno Ahi', url: 'https://pbs.twimg.com/profile_images/909918194701279232/UDem5J8J_400x400.jpg' },
      { name: 'Yandry Pozo', url: 'https://pbs.twimg.com/profile_images/876581578771636225/B8bT1qBI_400x400.jpg' }
    ];

    return (
      <Container>
        <Header />
        <Content>
          <List dataArray={photos}
            renderRow={(item) =>
              <ImageItem item={item} />
            }>
          </List>
        </Content>
      </Container>
    );
  }
}

class ImageItem extends Component {
  constructor(props) {
    super(props);
    this.state = { selected: false };
  }

  render() {
    const { item } = this.props;

    return (
      <ListItem
        onPress={() => {
          this.setState({ selected: !this.state.selected })
        }}
        style={this.state.selected ? styles.itemListSelected : styles.itemList}>

        <Thumbnail square size={180} source={{ uri: item.url }} />
        <Body>
          <Text>{item.name}</Text>
        </Body >
      </ListItem>
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