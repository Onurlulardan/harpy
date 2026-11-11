package apidocs

const referenceSignInCSS = `
html, body { height: 100%; margin: 0; }
body { display: flex; flex-direction: column; }
#harpy-signin {
  align-items: end; background: #111827; border-bottom: 1px solid #374151;
  color: #e5e7eb; display: flex; flex-wrap: wrap; font: 13px/1.4 system-ui, sans-serif;
  gap: 8px; padding: 10px 16px; position: sticky; top: 0; z-index: 20;
}
#harpy-signin label { display: grid; font-size: 11px; gap: 4px; }
#harpy-signin input, #harpy-signin button {
  background: #0b1020; border: 1px solid #4b5563; border-radius: 6px;
  color: inherit; padding: 6px 8px;
}
#harpy-signin button { background: #e5e7eb; color: #111827; cursor: pointer; font-weight: 700; }
#harpy-signin p { color: #9ca3af; flex-basis: 100%; margin: 0; }
#app { flex: 1; min-height: 0; }
`

const referenceSignInHeadJS = `
(function () {
  var tokenKey = "harpy-docs-access-token";
  var installed = false;
  Object.defineProperty(window, "Scalar", {
    configurable: true,
    set: function (value) {
      if (value && typeof value.createApiReference === "function" && !installed) {
        installed = true;
        var original = value.createApiReference.bind(value);
        value.createApiReference = function (selector, config) {
          var token = sessionStorage.getItem(tokenKey) || "";
          var authentication = Object.assign({}, config.authentication);
          var schemes = Object.assign({}, authentication.securitySchemes);
          schemes.bearerAuth = {type: "http", scheme: "bearer", token: token};
          authentication.securitySchemes = schemes;
          authentication.preferredSecurityScheme = "bearerAuth";
          config.authentication = authentication;
          config.persistAuth = true;
          return original(selector, config);
        };
      }
      Object.defineProperty(window, "Scalar", {
        configurable: true, writable: true, value: value
      });
    }
  });
})();
`

const referenceSignInBodyJS = `
(function () {
  var tokenKey = "harpy-docs-access-token";
  var bar = document.createElement("form");
  bar.id = "harpy-signin";
  bar.innerHTML = ''
    + '<label>Tenant<input name="tenant" value="harpy" autocomplete="off" required></label>'
    + '<label>E-posta<input name="email" type="email" value="superadmin@harpy.io" autocomplete="username" required></label>'
    + '<label>Parola<input name="password" type="password" autocomplete="current-password" required></label>'
    + '<label>Kalıcı parola<input name="next" type="password" autocomplete="new-password" placeholder="yalnızca ilk giriş"></label>'
    + '<button type="submit">Giriş yap</button>'
    + '<button type="button" data-clear>Tokeni sil</button>'
    + '<p data-status>Origin yazmana gerek yok. İlk paroladan sonra kalıcı parola en az 15 karakter olmalı ve içinde harpy geçmemeli.</p>';
  document.body.prepend(bar);
  var status = bar.querySelector("[data-status]");
  if (sessionStorage.getItem(tokenKey)) {
    status.textContent = "Access token bu sekmede duruyor. Sonraki istekler Authorization ile gider.";
  }
  bar.querySelector("[data-clear]").addEventListener("click", function () {
    sessionStorage.removeItem(tokenKey);
    location.reload();
  });
  bar.addEventListener("submit", function (event) {
    event.preventDefault();
    signIn(bar, status).catch(function (error) {
      status.textContent = error && error.message ? error.message : "Giriş tamamlanamadı";
    });
  });

  function signIn(form, status) {
    var tenant = form.tenant.value.trim();
    var email = form.email.value.trim();
    var password = form.password.value;
    var next = form.next.value;
    status.textContent = "Giriş kontrol ediliyor…";
    return postJSON("/api/auth/login", {
      tenant_slug: tenant, email: email, password: password
    }).then(function (result) {
      form.password.value = "";
      if (result.body && result.body.token) return keep(result.body.token);
      if (result.body && result.body.password_change_required) {
        if (!next) {
          status.textContent = "Geçici parola kabul edildi. Kalıcı parolayı yazıp yeniden gir.";
          return;
        }
        return complete(tenant, next, status);
      }
      fail(result);
    });
  }

  function complete(tenant, next, status) {
    var csrf = readCookie("__Host-harpy-initial-csrf");
    if (!csrf) throw new Error("Parola değiştirme çerezi okunamadı");
    return postJSON("/api/auth/initial-password", {
      tenant_slug: tenant, new_password: next
    }, {"x-csrf-token": csrf}).then(function (result) {
      if (result.body && result.body.token) return keep(result.body.token);
      fail(result);
    });
  }

  function postJSON(path, body, extra) {
    var headers = {"content-type": "application/json", accept: "application/json"};
    if (extra) Object.keys(extra).forEach(function (key) { headers[key] = extra[key]; });
    return fetch(path, {
      method: "POST", credentials: "same-origin", headers: headers,
      body: JSON.stringify(body)
    }).then(function (response) {
      return response.json().catch(function () { return {}; }).then(function (body) {
        return {ok: response.ok, status: response.status, body: body};
      });
    });
  }

  function keep(token) {
    sessionStorage.setItem(tokenKey, token);
    location.reload();
  }

  function fail(result) {
    var error = result.body && result.body.error;
    var code = error && error.code ? error.code : "HTTP " + result.status;
    var message = error && error.message ? error.message : "Giriş tamamlanamadı";
    throw new Error(code + ": " + message);
  }

  function readCookie(name) {
    var prefix = name + "=";
    var parts = document.cookie ? document.cookie.split("; ") : [];
    for (var index = 0; index < parts.length; index += 1) {
      if (parts[index].indexOf(prefix) === 0) {
        return decodeURIComponent(parts[index].slice(prefix.length));
      }
    }
    return "";
  }
})();
`
