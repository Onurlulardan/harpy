package main

const invitationGraphiteHTML = `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8">` +
	`<meta name="viewport" content="width=device-width,initial-scale=1">` +
	`<title>ForgeCRUD</title></head><body style="margin:0;background:#f4f4f5;` +
	`font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;color:#3f3f46">` +
	`<div style="max-width:560px;margin:32px auto;background:#fff;padding:40px;` +
	`border-radius:8px;border:1px solid #e4e4e7">` +
	`<div style="text-align:center;font-size:22px;font-weight:700;color:#18181b">ForgeCRUD</div>` +
	`<h1 style="margin:24px 0 12px;font-size:22px;color:#18181b;text-align:center">You're invited</h1>` +
	`<p style="margin:0 0 24px;text-align:center;line-height:1.6">` +
	`Join your ForgeCRUD workspace.</p>` +
	`<p style="text-align:center;margin:0">` +
	`<a href="{{INVITATION_URL}}" style="display:inline-block;background:#18181b;color:#fff;` +
	`text-decoration:none;padding:12px 28px;border-radius:6px;font-weight:600">` +
	`Accept invitation</a></p>` +
	`<p style="margin:28px 0 0;font-size:13px;color:#71717a;text-align:center">` +
	`If you did not expect this email, you can ignore it.</p></div></body></html>`
