<script>
export default {
	data() { return { conversations: [], errormsg: null, timer: null, search: "", users: [] } },
	methods: {
		async refresh() {
			this.errormsg = null;
			try {
				let response = await this.$axios.get("/conversations");
				this.conversations = response.data.conversations;
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async searchUsers() {
			this.errormsg = null;
			try {
				let response = await this.$axios.get("/users", { params: { username: this.search } });
				this.users = response.data.users;
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async startChat(userId) {
			this.errormsg = null;
			try {
				let response = await this.$axios.post("/conversations", { userId: userId });
				this.openChat(response.data.id);
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		photoUrl(photo) { return __API_URL__ + photo; },
		openChat(id) { this.$router.push("/conversations/" + id); },
	},
	mounted() {
		this.refresh();
		this.timer = setInterval(this.refresh, 2000);
	},
	beforeUnmount() { clearInterval(this.timer); },
}
</script>

<template>
	<div class="mt-3">
		<h1 class="h2 mb-3">Chats</h1>
		<ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>
		<div class="input-group mb-2" style="max-width: 400px">
			<input v-model="search" class="form-control" placeholder="Search users" @keyup.enter="searchUsers">
			<button class="btn btn-outline-primary" @click="searchUsers">Search</button>
		</div>
		<ul class="list-group mb-4" style="max-width: 400px">
			<li v-for="u in users" :key="u.id" class="list-group-item list-group-item-action" @click="startChat(u.id)">{{ u.name }}</li>
		</ul>
		<p v-if="conversations.length == 0">No conversations yet</p>
		<ul class="list-group">
			<li v-for="c in conversations" :key="c.id" class="list-group-item list-group-item-action" @click="openChat(c.id)">
				<img v-if="c.photo" :src="photoUrl(c.photo)" width="40" height="40" class="rounded-circle me-2">
				<b>{{ c.name }}</b>
				<span v-if="c.isGroup" class="badge bg-secondary ms-2">group</span>
				<div v-if="c.lastMessage" class="text-muted small">
					{{ new Date(c.lastMessage.timestamp).toLocaleString() }} -
					<svg v-if="c.lastMessage.isPhoto" class="feather"><use href="/feather-sprite-v4.29.0.svg#image"/></svg>
					{{ c.lastMessage.preview }}
				</div>
			</li>
		</ul>
	</div>
</template>

<style>
</style>
