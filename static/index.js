window.addEventListener("load", function (event) {
  const author = localStorage.getItem("author")
  document.getElementById("author").value = author
  document.getElementById("message").focus()
})

document
  .getElementById("addPostForm")
  .addEventListener("submit", function (event) {
    event.preventDefault()

    const author = document.getElementById("author").value
    const message = document.getElementById("message").value

    fetch("/api/v1/posts", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ author: author, message: message }),
    })
      .then((response) => {
        if (response.ok) {
          localStorage.setItem("author", author)
          window.location.reload()
        } else {
          alert("Failed to add post")
        }
      })
      .catch((error) => {
        console.error("Failed to make POST request:", error)
      })
  })
