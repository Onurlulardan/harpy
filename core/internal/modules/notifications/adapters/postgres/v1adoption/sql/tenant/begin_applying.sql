UPDATE notifications_adoption_runs
SET lifecycle_state = 'applying', row_version = row_version + 1
WHERE id = $1::uuid AND lifecycle_state = 'planned';
