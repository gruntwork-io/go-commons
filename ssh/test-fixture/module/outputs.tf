output "public_instance_ip" {
  value = aws_instance.example_public.public_ip
}

output "public_instance_private_ip" {
  value = aws_instance.example_public.private_ip
}

output "private_instance_ip" {
  value = aws_instance.example_private.private_ip
}

output "private_restricted_instance_ip" {
  value = aws_instance.example_private_restricted.private_ip
}
