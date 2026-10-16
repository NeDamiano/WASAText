<script>
export default {
	data() { return { conv: null, text: "", image: null, replyTo: null, reacting: null, forwarding: null, chats: [], emojis: ["👍", "❤️", "😂", "😮", "😢"], errormsg: null, timer: null, me: Number(sessionStorage.getItem("token")) } },
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
				this.image = null;
				this.replyTo = null;
				this.$refs.file.value = "";
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
				this.chats = response.data.conversations;
				this.forwarding = m.id;
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async forward(m, destinationId) {
			this.errormsg = null;
			try {
				await this.$axios.post("/conversations/" + this.$route.params.id + "/messages/" + m.id + "/forwards", { destinationId: destinationId });
				this.forwarding = null;
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		pickImage(e) { this.image = e.target.files[0]; },
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
		<ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>
		<div v-if="conv">
			<h1 class="h2 mb-3">
				<img v-if="conv.photo" :src="photoUrl(conv.photo)" width="40" height="40" class="rounded-circle me-2">
				{{ conv.name }}
			</h1>
			<div v-if="replyTo" class="alert alert-secondary py-1 mb-0 small text-truncate">
				<button class="btn-close float-end" @click="replyTo = null"></button>
				Reply to {{ replyTo.sender.name }}:
				<svg v-if="replyTo.photo" class="feather"><use href="/feather-sprite-v4.29.0.svg#image"/></svg>
				{{ replyTo.text }}
			</div>
			<div class="input-group mb-3">
				<input v-model="text" class="form-control" placeholder="Message" @keyup.enter="send">
				<label class="btn btn-outline-secondary">
					<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#paperclip"/></svg>
					{{ image ? image.name : "" }}
					<input ref="file" type="file" accept="image/*" hidden @change="pickImage">
				</label>
				<button class="btn btn-primary" @click="send">Send</button>
			</div>
			<div v-for="m in conv.messages" :key="m.id" class="card mb-2" :class="m.sender.id == me ? 'ms-auto bg-light' : ''" style="max-width: 70%; width: fit-content">
				<div class="card-body p-2">
					<div class="fw-bold">{{ m.sender.name }}</div>
					<div v-if="m.forwarded" class="small text-muted fst-italic">Forwarded</div>
					<div v-if="m.replyTo && original(m.replyTo)" class="border-start border-3 ps-2 mb-1 small text-muted text-truncate">
						<svg v-if="original(m.replyTo).photo" class="feather"><use href="/feather-sprite-v4.29.0.svg#image"/></svg>
						{{ original(m.replyTo).text }}
					</div>
					<img v-if="m.photo" :src="photoUrl(m.photo)" class="img-fluid mb-1" style="max-height: 400px">
					<div v-if="m.text" class="fs-6">{{ m.text }}</div>
					<div v-if="m.comments">
						<span v-for="c in m.comments" :key="c.id" class="badge text-bg-light border fw-normal fs-6 me-1">
							{{ c.emoji }} {{ c.sender.name }}
							<button v-if="c.sender.id == me" class="btn-close ms-1" style="font-size: 0.5rem" @click="unreact(m, c)"></button>
						</span>
					</div>
					<div class="text-muted text-end">
						<button class="btn btn-link p-0 me-2" @click="reacting = reacting == m.id ? null : m.id">React</button>
						<button class="btn btn-link p-0 me-2" @click="replyTo = m">Reply</button>
						<button class="btn btn-link p-0 me-2" @click="startForward(m)">Forward</button>
						<button v-if="m.sender.id == me" class="btn btn-link text-danger p-0 me-2" @click="deleteMessage(m)">Delete</button>
						{{ new Date(m.timestamp).toLocaleString() }}
						<span v-if="m.sender.id == me">{{ ticks(m.status) }}</span>
					</div>
					<div v-if="reacting == m.id" class="text-end">
						<button v-for="e in emojis" :key="e" class="btn btn-sm" @click="react(m, e)">{{ e }}</button>
					</div>
					<div v-if="forwarding == m.id" class="mt-2">
						<div class="small text-muted">
							Forward to:
							<button class="btn-close float-end" @click="forwarding = null"></button>
						</div>
						<ul class="list-group">
							<li v-for="c in chats" :key="c.id" class="list-group-item list-group-item-action" @click="forward(m, c.id)">{{ c.name }}</li>
						</ul>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<style>
</style>
