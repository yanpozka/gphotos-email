import { Component } from 'react';
import { Body, Container, List, ListItem, Text, Thumbnail } from 'native-base';


export default class PhotosList extends Component {
	render() {
		var photos = [
			{ name: 'Uno Ahi', url: 'https://pbs.twimg.com/profile_images/909918194701279232/UDem5J8J_400x400.jpg' },
			{ name: 'Yandry Pozo', url: 'https://pbs.twimg.com/profile_images/876581578771636225/B8bT1qBI_400x400.jpg' }
		];

		return (
			<Container>
				<List dataArray={photos}
					renderRow={(item) =>
						<ImageItem item={item} />
					}>
				</List>
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