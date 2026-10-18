<script>
export default {
	data() { return { conv: null, text: "", image: null, replyTo: null, reacting: null, forwarding: null, chats: [], users: [], emojis: ["👍", "❤️", "😂", "🙂", "😮", "😢"], errormsg: null, timer: null, me: Number(sessionStorage.getItem("token")) } },
	methods: {
		async refresh() {
			try {
				let response = await this.$axios.get("/conversations/" + this.$route.params.id);
				this.conv = response.data;
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async send() {
			this.errormsg = null;
			let form = new FormData();
			if (this.text) form.append("text", this.text);
			if (this.image) form.append("image", this.image);
			if (this.replyTo) form.append("replyTo", this.replyTo.id);
			try {
				await this.$axios.post("/conversations/" + this.$route.params.id + "/messages", form);
				this.text = "";
				this.replyTo = null;
				this.clearImage();
				this.refresh();
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async react(m, emoji) {
			this.errormsg = null;
			try {
				await this.$axios.post("/conversations/" + this.$route.params.id + "/messages/" + m.id + "/comments", { emoji: emoji });
				this.reacting = null;
				this.refresh();
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async unreact(m, c) {
			this.errormsg = null;
			try {
				await this.$axios.delete("/conversations/" + this.$route.params.id + "/messages/" + m.id + "/comments/" + c.id);
				this.refresh();
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async deleteMessage(m) {
			this.errormsg = null;
			try {
				await this.$axios.delete("/conversations/" + this.$route.params.id + "/messages/" + m.id);
				this.refresh();
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async startForward(m) {
			this.errormsg = null;
			try {
				let response = await this.$axios.get("/conversations");
				this.chats = response.data.conversations.filter(x => x.isGroup);
				response = await this.$axios.get("/users");
				this.users = response.data.users.filter(x => x.id != this.me);
				this.forwarding = m;
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async forward(destinationId) {
			this.errormsg = null;
			try {
				await this.$axios.post("/conversations/" + this.$route.params.id + "/messages/" + this.forwarding.id + "/forwards", { destinationId: destinationId });
				this.forwarding = null;
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async forwardToUser(userId) {
			this.errormsg = null;
			try {
				let response = await this.$axios.post("/conversations", { userId: userId });
				this.forward(response.data.id);
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		pickImage(e) { this.image = e.target.files[0]; },
		clearImage() { this.image = null; this.$refs.file.value = ""; },
		original(id) { return this.conv.messages.find(x => x.id == id); },
		photoUrl(photo) { return __API_URL__ + photo; },
		ticks(status) { return status == "read" ? "✓✓" : status == "received" ? "✓" : ""; },
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
		<div class="position-fixed top-0 end-0 m-3" style="z-index: 1060">
			<ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>
		</div>
		<div v-if="forwarding" class="modal d-block" style="background: rgba(0, 0, 0, 0.5)">
			<div class="modal-dialog">
				<div class="modal-content">
					<div class="modal-header">
						<h5 class="modal-title">Forward to</h5>
						<button class="btn-close" @click="forwarding = null"></button>
					</div>
					<ul class="list-group list-group-flush">
						<li v-for="c in chats" :key="c.id" class="list-group-item list-group-item-action" @click="forward(c.id)">{{ c.name }}</li>
						<li v-for="u in users" :key="u.id" class="list-group-item list-group-item-action" @click="forwardToUser(u.id)">{{ u.name }}</li>
					</ul>
				</div>
			</div>
		</div>
		<div v-if="conv" class="d-flex flex-column" style="height: calc(100vh - 80px)">
			<h1 class="h2 mb-3">
				<img v-if="conv.photo" :src="photoUrl(conv.photo)" width="40" height="40" class="rounded-circle me-2">
				{{ conv.name }}
				<button v-if="conv.isGroup" class="btn btn-outline-secondary ms-2" @click="$router.push('/groups/' + conv.id)">Group settings</button>
			</h1>
			<div class="flex-grow-1 overflow-auto d-flex flex-column-reverse mb-2 pe-3">
				<div v-for="m in conv.messages" :key="m.id" class="card mb-2" :class="m.sender.id == me ? 'ms-auto bg-success bg-opacity-25' : ''" style="max-width: 70%; width: fit-content">
					<div class="card-body p-2">
						<div v-if="m.sender.id != me" class="fw-bold">{{ m.sender.name }}</div>
						<div v-if="m.forwarded" class="small text-muted fst-italic">Forwarded</div>
						<div v-if="m.replyTo && original(m.replyTo)" class="border-start border-3 ps-2 mb-1 small text-muted text-truncate">
							<svg v-if="original(m.replyTo).photo" class="feather"><use href="/feather-sprite-v4.29.0.svg#image"/></svg>
							{{ original(m.replyTo).text }}
						</div>
						<img v-if="m.photo" :src="photoUrl(m.photo)" class="img-fluid mb-1" style="max-height: 400px">
						<div v-if="m.text">{{ m.text }}</div>
						<div v-if="m.comments">
							<span v-for="c in m.comments" :key="c.id" class="badge text-bg-light border fw-normal fs-6 me-1">
								{{ c.emoji }} {{ c.sender.name }}
								<button v-if="c.sender.id == me" class="btn-close ms-1" style="font-size: 0.5rem" @click="unreact(m, c)"></button>
							</span>
						</div>
						<div class="text-muted text-end">
							<div class="btn-group btn-group-sm me-2">
								<button class="btn btn-outline-dark" title="React" @click="reacting = reacting == m.id ? null : m.id">
									<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#smile"/></svg>
								</button>
								<button class="btn btn-outline-dark" title="Reply" @click="replyTo = m">
									<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#corner-up-left"/></svg>
								</button>
								<button class="btn btn-outline-dark" title="Forward" @click="startForward(m)">
									<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#corner-up-right"/></svg>
								</button>
								<button v-if="m.sender.id == me" class="btn btn-outline-danger" title="Delete" @click="deleteMessage(m)">
									<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#trash-2"/></svg>
								</button>
							</div>
							{{ new Date(m.timestamp).toLocaleString() }}
							<span v-if="m.sender.id == me">{{ ticks(m.status) }}</span>
						</div>
						<div v-if="reacting == m.id" class="text-end">
							<button v-for="e in emojis" :key="e" class="btn btn-sm" @click="react(m, e)">{{ e }}</button>
						</div>
					</div>
				</div>
			</div>
			<div v-if="replyTo" class="alert alert-secondary py-1 mb-0 small text-truncate">
				<button class="btn-close float-end" @click="replyTo = null"></button>
				Reply to {{ replyTo.sender.name }}:
				<svg v-if="replyTo.photo" class="feather"><use href="/feather-sprite-v4.29.0.svg#image"/></svg>
				{{ replyTo.text }}
			</div>
			<div class="input-group">
				<input v-model="text" class="form-control" placeholder="Message" @keyup.enter="send">
				<label class="btn btn-outline-secondary">
					<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#paperclip"/></svg>
					{{ image ? image.name : "" }}
					<input ref="file" type="file" accept="image/*" hidden @change="pickImage">
				</label>
				<button v-if="image" class="btn btn-outline-secondary" title="Remove photo" @click="clearImage">
					<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#x"/></svg>
				</button>
				<button class="btn btn-primary" @click="send">Send</button>
			</div>
		</div>
	</div>
</template>

<style>
</style>
